package treap

import (
	"fmt"
	"math/rand"
	"testing"
)

// TestInsertarYBuscar verifica que los elementos se inserten y recuperen correctamente.
func TestInsertarYBuscar(t *testing.T) {
	treap := NuevoTreap()

	// Insertar datos de prueba
	treap.Insertar("7750001", "Gatorade")
	treap.Insertar("7750002", "Doritos")

	// Caso 1: Buscar elemento existente
	val, encontrado := treap.Buscar("7750001")
	if !encontrado {
		t.Error("Se esperaba encontrar el producto '7750001'")
	}
	if val.(string) != "Gatorade" {
		t.Errorf("Se esperaba 'Gatorade', pero se obtuvo '%v'", val)
	}

	// Caso 2: Buscar elemento inexistente
	_, encontrado = treap.Buscar("9999999")
	if encontrado {
		t.Error("No se esperaba encontrar el producto inexistente '9999999'")
	}
}

// TestEliminar verifica que el nodo se remueva manteniendo la integridad del árbol.
func TestEliminar(t *testing.T) {
	treap := NuevoTreap()

	treap.Insertar("7750003", "Agua San Luis")

	// Eliminar el elemento
	treap.Eliminar("7750003")

	// Verificar que ya no exista
	_, encontrado := treap.Buscar("7750003")
	if encontrado {
		t.Error("El producto '7750003' no debió ser encontrado tras su eliminación")
	}
}

// TestListarTodosInOrder comprueba que el catálogo devuelva los datos ordenados alfabéticamente.
func TestListarTodosInOrder(t *testing.T) {
	treap := NuevoTreap()

	// Insertar de forma desordenada
	treap.Insertar("C", "Producto C")
	treap.Insertar("A", "Producto A")
	treap.Insertar("B", "Producto B")

	productos := treap.ListarTodos()

	if len(productos) != 3 {
		t.Fatalf("Se esperaban 3 productos, se obtuvieron %d", len(productos))
	}

	// Al ser in-orden, deben salir estrictamente en secuencia A -> B -> C
	if productos[0].(string) != "Producto A" ||
		productos[1].(string) != "Producto B" ||
		productos[2].(string) != "Producto C" {
		t.Error("El recorrido in-orden no devolvió los elementos en el orden alfabético correcto de sus claves")
	}
}

// TestPropiedadHeap garantiza que las rotaciones mantienen el orden de prioridad i.i.d.
func TestPropiedadHeap(t *testing.T) {
	treap := NuevoTreap()

	// Insertamos múltiples elementos para forzar rotaciones pesadas
	treap.Insertar("E", "Prod 5")
	treap.Insertar("D", "Prod 4")
	treap.Insertar("C", "Prod 3")
	treap.Insertar("B", "Prod 2")
	treap.Insertar("A", "Prod 1")

	if !validarHeap(treap.Raiz) {
		t.Error("Se violó la propiedad de Heap: un nodo hijo tiene mayor prioridad que su padre")
	}
}

func validarHeap(nodo *Nodo) bool {
	if nodo == nil {
		return true
	}
	if nodo.Izq != nil && nodo.Izq.Prioridad > nodo.Prioridad {
		return false
	}
	if nodo.Der != nil && nodo.Der.Prioridad > nodo.Prioridad {
		return false
	}
	return validarHeap(nodo.Izq) && validarHeap(nodo.Der)
}

// BENCHMARKS

var dataSet []string
var tBench *Treap

func init() {
	// 10k elementos como tamaño base representativo
	const size = 10000
	dataSet = make([]string, size)

	r := rand.New(rand.NewSource(42))
	for i := range size {
		dataSet[i] = fmt.Sprintf("KEY-%06d", r.Intn(size*10))
	}

	tBench = NuevoTreap()
	for _, k := range dataSet {
		tBench.Insertar(k, "item-data")
	}
}

// mide el rendimiento de inserciones individuales en un árbol limpio
func BenchmarkInsertar(b *testing.B) {
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		treap := NuevoTreap()
		clave := dataSet[i%len(dataSet)]
		treap.Insertar(clave, "valor")
	}
}

// mide el costo amortizado de poblar un único treap de forma masiva
func BenchmarkInsertarSecuencial(b *testing.B) {
	treap := NuevoTreap()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clave := dataSet[i%len(dataSet)]
		treap.Insertar(clave, "valor")
	}
}

// mide el tiempo de búsqueda (hit/miss) en un treap ya poblado
func BenchmarkBuscar(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		clave := dataSet[i%len(dataSet)]
		_, _ = tBench.Buscar(clave)
	}
}

// para evitar vaciar el árbol global, clonamos/reconstruimos el estado en baches controlados
func BenchmarkEliminar(b *testing.B) {
	for i := 0; i < b.N; i++ {
		treap := NuevoTreap()
		for j := range 100 {
			treap.Insertar(dataSet[j], "data")
		}
		claveAEliminar := dataSet[i%100]

		treap.Eliminar(claveAEliminar)
	}
}
