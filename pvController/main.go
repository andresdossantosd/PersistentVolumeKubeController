// Es necesario definir un paquete, llamaremos main
// Go no es OOP
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// utilizamos flag para parsear los parametros de entrada de entrada del programa
	kubeconfig := flag.String("kubeconfig", "/home/master/.kube/config", "Ruta absoluta al .config de kubectl")
	flag.Parse()
	// construye configuracion del cliente --> clientcmd.BuildConfigFromFlags(masterUrl string, kubeconfigPath string) builds configs from a master url or a kubeconfig filepath
	// creamos la configuración del cliente partiendo del fichero del kubeconfig (OJO: solo la configuración)
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatalf("FAIL: configuracion del ~/.kube/config --> %s", err.Error())
	}

	// creamos cliente con la configuración `config` que creamos previamente para conectarnos al api-server del Cluster --> kubernetes.NewForConfig(c *rest.Config) (*kubernetes.Clientset, error) NewForConfig creates a new Clientset for the given config
	cliente, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatalf("FAIL: creando el cliente que se conectara a la API --> %s", err.Error())
	}

	// Creamos un watcher para los Persistent volumes. Recibira/listara los cambios en el api-server sobre recursos de PV's (es como usar ?watch=true)
	watcher := cache.NewListWatchFromClient(
		cliente.CoreV1().RESTClient(), // utilizar cliente REST de Kubernetes v1
		"persistentvolumes",           // para observar persistent volumes
		metav1.NamespaceAll,           // observa todos los namespaces
		fields.Everything(),           // Everything() returns a selector that matches all fields. --> Package fields implements a simple field system, parsing and matching selectors with sets of fields
	)

	// Definimos el informador del controlador: quien se va a encargar de escuchar eventos provenientes del watcher creado
	// NewInformer returns a Store and a controller for populating the store while also providing event notifications.
	_, controlador := cache.NewInformer(
		watcher,
		&v1.PersistentVolume{},
		time.Second*0,
		cache.ResourceEventHandlerFuncs{
			//--> OJO, AQUI CREAMOS EL HANDLER DE LOS EVENTOS, AQUI SE QUEDARA ESCUCHANDO
			AddFunc: func(obj interface{}) {
				// esta funcion es la que se encarga de manejar el evento de agregar un PV
				pv := obj.(*v1.PersistentVolume)                       // de objeto que se esta observando a persisten Volume
				fmt.Printf("Persistent Volume created: %s\n", pv.Name) // Imprime el nombre del PV creado

				// PV no tienen namespace, por ello le damos el default
				namespace := pv.Namespace
				if namespace == "" {
					namespace = "default"
				}
				// Creamos un evento para el PV creado
				// Definimos un  nuevo evento de kubernetes
				// &v1.Event --> Event es un struct
				event := &v1.Event{
					// definimos los metadatos del evento, indicamos prefijo y el namespace en el que se crea
					ObjectMeta: metav1.ObjectMeta{
						GenerateName: "pv-event-",
						Namespace:    namespace,
					},
					// Indicamos el objeto involucrado en el evento y su namespace
					InvolvedObject: v1.ObjectReference{
						Kind:      "PersistentVolume",
						Name:      pv.Name,
						Namespace: namespace,
					},
					// mensaje del evento
					Message: "Persistent Volume created: " + pv.Name,
					// la fuente del evento, y especificamos dentro el componente que genera el evento, es decir, EL CONTROLADOR DEL PERSISTENT VOLUME:
					// EventSource contains information for an event.
					Source: v1.EventSource{
						Component: "pv-controller",
					},
					// TIMESTAMP DE CREACION Y ULTIMO EVENTO
					FirstTimestamp: metav1.Now(),
					LastTimestamp:  metav1.Now(),
					// tipo del evento --> Type of this event (Normal, Warning), new types could be added in the future +optional
					Type: "Normal",
				}
				// Crea el evento en el cluster de kubernetes
				_, err := cliente.CoreV1().Events(namespace).Create(context.TODO(), event, metav1.CreateOptions{})
				if err != nil {
					log.Fatalf("FAIL: no se ha creado el evento %s", err.Error())
				}
			},
		},
	)
	// necesario para detener el controlador
	stop := make(chan struct{})
	defer close(stop)
	go controlador.Run(stop)
	select {}
}
