package main

import (
	"fmt"
	"strconv"
	"time"

	"k8s.io/client-go/util/workqueue"
)

// =======================================================================================================================================================================
// ==================================================== Documentación: Colas en K8s con client-go ========================================================================
// =======================================================================================================================================================================

// Paquete --> k8s.io/client-go/util/workqueue: provee interfaz de implementacion de colas
// 					Package workqueue provides a simple queue that supports the following features:

// 					-	Fair: items processed in the order in which they are added.
// 					-	Stingy: a single item will not be processed multiple times concurrently, and if an item is added multiple times before it can be processed,
//						it will only be processed once.
// 					-	Multiple consumers and producers. In particular, it is allowed for an item to be reenqueued while it is being processed.
// 					-	Shutdown notifications.

// La libreria de client-go para comunicarnos con el kube-apiserver, utiliza de estas colas
// Estas colas son threads safe y estan implementadas con canales de go (go Channels)

// Ejercicio a realizar: crear un sender y receiver para ver como interactuar con channels
// y comprobar la sincronizacion utilizando canales entre diferentes hilos

// ¿ Como funciona las colas de client-go ? https://leftasexercise.com/2019/07/08/understanding-kubernetes-controllers-part-i-queues-and-the-core-controller-loop/ en el apartado de Queues and concurrency
//		Pero en resumidas, se implementan con 3 sets, un array con los elementos, un Dirty Set (elmentos no procesados) y un Processing Set
//			-	Al ejecutar Add(), lo inserta en el array y dirty set
//			-	Al ejecutar Get(), lo saca del dirty set y lo envia al processing set, y saca el primer elemento del array (Pop)
//			-	Al ejecutar Done(), lo saca del processing set y si entre medias, alguien metio ese mismo objeto (hizo un add), se vuelve a insertar en la cola

// Rellenar la cola
// Las operaciones de workqueue.Interface struct son:
// 		type Interface interface {
// 			Add(item interface{}) --> Add marks item as needing processing.
// 			Len() int
// 			Get() (item interface{}, shutdown bool) --> Done marks item as done processing, and if it has been marked as dirty again while it was being processed, it will be re-added to the queue for re-processing.
// 			Done(item interface{})
// 			ShutDown()
// 			ShutDownWithDrain()
// 			ShuttingDown() bool
// 		}

// =======================================================================================================================================================================

// Elementos a insertar en la cola

type element struct {
	index uint64
	name  string
}

func addToTheQueue(queue workqueue.Interface) {
	time.Sleep(10 * time.Second)
	for i := 0; i < 30; i++ {
		// Aqui cerramos, ilustramos que cierra y el ultimo elemento en ingresar en el 19
		// Inserta Nil al final
		if i == 20 {
			queue.ShutDown()
		}
		queue.Add(element{index: uint64(i), name: "Hi Im index " + strconv.Itoa(i)})
	}
}

func readFromQueue(queue workqueue.Interface, canalEspera chan int) {
	time.Sleep(5 * time.Second)
	for {
		// Get blocks until it can return an item to be processed.
		// If shutdown = true, the caller should end their goroutine.
		item, closed := queue.Get()
		e, _ := item.(element)
		if closed {
			fmt.Printf("Element : %s , does queue has been shutdown ? %t \n", e, closed)
			// envias al canal un valor para que ya no se quede esperando
			canalEspera <- -1
			break
		}
		fmt.Printf("Element : %s , does queue has been shutdown ? %t \n", e.name, closed)
		// You must call Done with item when you have finished processing it.
		queue.Done(item)
	}
}

func main() {
	canalEspera := make(chan int)
	thaQueue := workqueue.New()
	go addToTheQueue(thaQueue)
	go readFromQueue(thaQueue, canalEspera)
	// Esperamos a leer del canal, el cual se desbloquea al escribirle un valor (al canal)
	<-canalEspera
}
