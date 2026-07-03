Inventario Tambo — Treap (Randomized Search Trees)

Sistema de inventario para un mini-market (tambo), construido sobre un Treap (árbol binario de búsqueda que se auto-balancea mediante prioridades aleatorias) como índice en memoria, respaldado por SQLite para persistencia en disco. El backend está en Go y expone una API REST, el frontend es Vue 3 con componentes de Material Design 3 servidos sin bundler, y la estructura central del proyecto es el paquete internal/treap.

1. Instrucciones de ejecución

Requisitos previos:

Go 1.26.4 o superior, que es la versión declarada en el go.mod. Si se tiene una versión anterior instalada, alcanza con bajar el número de la directiva go en el go.mod al que corresponda, siempre que sea 1.21 o superior (por el uso de range sobre enteros en los benchmarks de treap_test.go).

Cgo habilitado, porque github.com/mattn/go-sqlite3 compila un binding a la librería C de SQLite. En Linux o WSL hace falta tener gcc instalado (build-essential en Debian/Ubuntu), en macOS las Command Line Tools de Xcode (xcode-select --install), y en Windows un toolchain como MinGW-w64 o TDM-GCC agregado al PATH. Si CGO_ENABLED quedó en 0 en el entorno, el build va a fallar con un error de tipo "Binary was compiled with 'CGO_ENABLED=0'"; para forzarlo alcanza con CGO_ENABLED=1 go run main.go.

No hace falta instalar un motor de base de datos aparte: SQLite es embebido y el archivo inventario.db se crea solo la primera vez que se levanta el servidor, si no existe, junto con la tabla inventario. Si en algún momento se quiere arrancar de cero, simplemente se borra ese archivo antes de correr el servidor.

El frontend no necesita build ni npm install: index.html importa Vue 3, Lit, tslib y los componentes M3E directo desde jsdelivr usando import maps, así que solo hace falta conexión a internet la primera vez que el navegador cachea esos recursos. Si se va a desplegar en un entorno sin salida a internet, esos mismos paquetes tendrían que empaquetarse localmente y ajustar las rutas del import map.

Un navegador moderno que soporte import maps y web components de forma nativa: versiones recientes de Chrome, Edge, Firefox o Safari. En navegadores viejos los componentes m3e- simplemente no van a registrar y la interfaz queda en blanco.

El puerto 8080 tiene que estar libre. Si ya está en uso por otro proceso, http.ListenAndServe va a devolver un error de bind y el programa termina con log.Fatal; en ese caso hay que liberar el puerto o cambiar el ":8080" en main.go por otro.

La estructura de carpetas esperada es:

proyecto-treap/
  go.mod
  go.sum
  main.go
  inventario.db (se genera solo, no versionar en git)
  internal/database/database.go
  internal/treap/treap.go
  internal/treap/treap_test.go
  frontend/index.html
  frontend/styles.css

El servidor sirve frontend/ como archivos estáticos con http.FileServer, así que la carpeta tiene que llamarse exactamente frontend y estar en el mismo directorio desde el que se ejecuta el binario o el go run, porque la ruta "./frontend" en main.go es relativa al directorio de trabajo actual.

Comandos de arranque:

Primero descargar las dependencias del módulo:

go mod tidy

Después levantar el backend:

go run main.go

Esto sirve la API y el frontend estático en el mismo proceso, sobre el puerto 8080. Al arrancar, el servidor inicializa inventario.db si no existe, carga todos los productos ya guardados en SQLite hacia el treap en memoria con CargarProductos, imprime en consola cuántos productos se cargaron, y deja el frontend disponible en http://localhost:8080/ y la API en http://localhost:8080/api/productos.

No hace falta compilar el frontend por separado: como es Vue 3 sin bundler, con abrir http://localhost:8080 en el navegador ya queda funcionando todo, incluyendo las tres secciones de la interfaz (Caja, Almacén y Catálogo).

Si se quiere un binario aparte para desplegar:

go build -o inventario-tambo .
./inventario-tambo

El binario resultante todavía necesita encontrar la carpeta frontend y poder crear o leer inventario.db en el directorio desde el que se lo ejecute, así que conviene copiar frontend/ junto al binario si se lo mueve a otro lado.

Para probar la API directamente sin pasar por la interfaz, con curl:

curl http://localhost:8080/api/productos
curl "http://localhost:8080/api/productos?codigo=7750001"
curl -X POST http://localhost:8080/api/productos -H "Content-Type: application/json" -d "{\"Codigo\":\"7750001\",\"Nombre\":\"Gatorade\",\"Precio\":6.5,\"Stock\":24}"
curl -X DELETE "http://localhost:8080/api/productos?codigo=7750001"

Ejecución de pruebas y rendimiento:

Para correr las pruebas unitarias del paquete treap (inserción, búsqueda, eliminación, recorrido in-order y validación de la propiedad de heap):

go test ./internal/treap/...

Con detalle de cada caso ejecutado:

go test -v ./internal/treap/...

Para los benchmarks (inserción individual sobre un treap limpio, inserción secuencial masiva sobre un mismo treap, búsqueda hit/miss, y eliminación en bloques de 100 elementos), con reporte de asignaciones de memoria por operación:

go test -bench=. -benchmem ./internal/treap/...

Si se quiere correr un benchmark puntual, por ejemplo solo búsqueda:

go test -bench=BenchmarkBuscar -benchmem ./internal/treap/...

Y para todo junto, pruebas y benchmarks en una sola pasada:

go test -v -bench=. -benchmem ./internal/treap/...

El archivo treap_test.go arma un dataset de 10.000 claves con semilla fija (rand.NewSource(42)) para que los benchmarks sean reproducibles entre corridas; los resultados de ns/op y B/op sirven como referencia relativa para justificar el comportamiento O(log n) esperado, comparando cómo escala el tiempo de búsqueda o inserción a medida que crece el tamaño del dataset de prueba.

2. Análisis de complejidad Big-O

Complejidad asintótica:

Inserción: O(log n) en el caso promedio, O(n) en el peor caso. Es una búsqueda binaria normal, descendiendo por Izq o Der según la comparación de claves, más como mucho una rotación por cada nivel del camino recorrido, para restaurar la propiedad de heap después de insertar.

Búsqueda: O(log n) en el caso promedio, O(n) en el peor caso. Es exactamente igual a la búsqueda de un BST común, sin ningún costo adicional por ser treap: no se tocan prioridades ni se rota nada.

Eliminación: O(log n) en el caso promedio, O(n) en el peor caso. Primero se ubica el nodo con una búsqueda binaria estándar, y si tiene dos hijos se lo va rotando hacia abajo, siempre hacia el lado de mayor prioridad, hasta que queda como hoja o con un solo hijo, momento en el que se lo desconecta directamente.

Listar todo (ListarTodos): O(n) siempre, sin mejor ni peor caso, porque el recorrido in-order visita los n nodos sin excepción y sin poder saltarse ninguno.

Rotación individual (rotarIzquierda, rotarDerecha): O(1), porque cada una solo reasigna tres punteros (x.Der, y.Izq y el retorno de la nueva raíz del subárbol), sin recorrer ni tocar los subárboles que cuelgan de esos nodos.

Espacio: O(n), un nodo por cada elemento guardado, cada uno con una clave, un valor, una prioridad float64 y dos punteros.

Sustentación teórica:

Cada nodo del treap combina dos propiedades simultáneas. Sobre la clave se cumple la propiedad de BST de siempre: todo lo que cuelga a la izquierda es menor y todo lo que cuelga a la derecha es mayor. Sobre la prioridad, que se genera al azar e independiente en CrearNodo con t.rnd.Float64() (un valor continuo uniforme en [0, 1)), se cumple la propiedad de max-heap: todo nodo tiene mayor prioridad que sus hijos, de manera que la raíz siempre termina siendo, de entre todos los nodos presentes, el de mayor prioridad.

Seidel y Aragon, en su paper de 1996 publicado en Algorithmica bajo el título Randomized Search Trees, demuestran que al asignar esas prioridades de forma aleatoria e independiente entre sí, la forma que termina teniendo el árbol resultante es equivalente, en distribución de probabilidad, a la que tendría un BST construido insertando esas mismas claves pero en un orden aleatorio uniforme, sin importar el orden real en que llegaron los datos al sistema. En otras palabras: el treap simula, mediante las prioridades, el mismo efecto de balance que se obtendría barajando al azar el orden de inserción, aunque las claves en la realidad hayan llegado en cualquier secuencia, incluso ya ordenadas.

Ese resultado es justo lo que necesita este proyecto: los productos se insertan en el treap en el orden en que llegan desde SQLite al arrancar el servidor (CargarProductos itera las filas en el orden que devuelve la consulta SQL, que no está garantizado que sea por código), o en el orden en que el operador los va dando de alta desde la pestaña de Almacén, un orden que en la práctica puede venir agrupado por proveedor, por fecha de carga, o de cualquier otra forma que no sea alfabética. Gracias a la aleatoriedad de las prioridades, ese orden de llegada deja de ser un problema para el balance del árbol: el treap se comporta, en esperanza, como si esos mismos códigos se hubieran insertado en orden aleatorio.

De esa equivalencia con un BST aleatorio salen las cotas esperadas de la tabla anterior. La altura esperada de un BST construido con n inserciones en orden aleatorio es del orden de 2 ln n, que es O(log n); como cada operación de búsqueda, inserción o eliminación recorre a lo sumo un camino desde la raíz hasta una hoja, su costo esperado hereda directamente esa cota de altura. El costo de las rotaciones no cambia esa cota: cada rotación individual es O(1), y la cantidad total de rotaciones necesarias para restaurar la propiedad de heap después de una inserción o eliminación es, en amortizado, también O(1) por operación, porque en promedio solo hace falta reacomodar un número constante de niveles cercanos al punto donde se insertó o eliminó el nodo, no todo el camino recorrido en la búsqueda.

El peor caso de O(n) sigue existiendo en teoría: pasaría si, por mala suerte, el generador aleatorio le asignara a las prioridades justo el orden más desfavorable posible, uno que haga que el treap degenere en una lista enlazada, algo equivalente a insertar datos ya ordenados en un BST sin ningún mecanismo de balanceo. Pero la probabilidad de que esa combinación exacta de prioridades ocurra cae de forma exponencial a medida que crece n, así que en la práctica, con miles de productos como en este inventario, es un caso que estadísticamente no aparece. Vale la pena notar también que las prioridades se generan con una semilla basada en time.Now().UnixNano() en NuevoTreap, lo que evita que corridas sucesivas del programa reproduzcan siempre el mismo árbol degenerado por coincidencia.

El test TestPropiedadHeap en treap_test.go valida justamente esto: después de insertar varios elementos, recorre todo el árbol con validarHeap y confirma que ningún nodo hijo tenga una prioridad mayor a la de su padre, es decir, que la propiedad de heap nunca se haya roto tras las rotaciones aplicadas durante las inserciones.

3. Detalles del proyecto

Justificación del caso de uso:

El treap funciona acá como índice vivo en memoria sobre el código de barras de cada producto, mientras SQLite se ocupa exclusivamente de la persistencia en disco; son dos representaciones del mismo inventario, optimizadas para cosas distintas.

Los datos no llegan ordenados: los códigos de barras se insertan en el orden en que el operador los registra desde la pestaña de Almacén, o en el orden en que ya estaban guardados en inventario.db al momento de arrancar el servidor, que puede venir agrupado por lotes de carga, por proveedor, o de cualquier forma que no tenga relación con el orden alfabético de los códigos. Un BST simple sin balancear se degradaría fácilmente ante secuencias así, sobre todo si en algún momento se cargan productos con códigos consecutivos de forma masiva. El treap resuelve esto sin necesitar reglas de coloreo como en un árbol Rojo-Negro ni factores de balance y rotaciones condicionadas por altura como en AVL: alcanza con comparar dos prioridades aleatorias en cada rotación.

La búsqueda en caja tiene que sentirse instantánea: el endpoint GET /api/productos?codigo=..., que se dispara cada vez que se escanea o se tipea un código en la pestaña Caja, se resuelve con Treap.Buscar, en O(log n) esperado. Con el inventario típico de un mini-market, que rara vez supera unos pocos miles de códigos distintos, esa cota se traduce en prácticamente ninguna demora perceptible durante el flujo de venta, incluso en un dataset que fuera creciendo con el tiempo.

El catálogo tiene que salir ordenado alfabéticamente sin trabajo extra: la pestaña Catálogo llama a GET /api/productos, que por dentro usa Treap.ListarTodos, un recorrido in-order en O(n). Como la propiedad de BST del treap ya mantiene las claves ordenadas en todo momento, listar el inventario completo en orden alfabético no necesita ningún ORDER BY adicional contra SQLite ni una estructura de datos aparte: es simplemente una consecuencia de cómo está organizado el árbol en memoria.

Las escrituras son dobles pero quedan desacopladas y consistentes: tanto el alta o edición de un producto (POST) como su baja (DELETE) siguen el mismo patrón en manejarProductos, primero se persiste el cambio en SQLite con GuardarProducto o EliminarProducto, usando un UPSERT sobre la clave primaria codigo para evitar duplicados o inconsistencias si se reenvía el mismo código dos veces, y solo si esa escritura en disco sale bien se replica la operación en el treap en memoria con srv.t.Insertar o srv.t.Eliminar. De esa manera, si en algún momento falla la escritura en SQLite, el treap en memoria no queda desincronizado con lo que realmente hay persistido en disco.

Sobre las limitaciones actuales, vale la pena tenerlas presentes de cara a una eventual sustentación o a una siguiente iteración del proyecto. El Treap no tiene ningún mecanismo de sincronización interno (no usa sync.Mutex ni sync.RWMutex sobre Raiz), y como net/http atiende cada request entrante en su propia goroutine, dos escrituras concurrentes sobre el mismo servidor podrían en teoría generar una condición de carrera sobre el árbol en memoria; en el volumen de tráfico esperado para un solo punto de venta esto es poco probable que se manifieste, pero sería el primer punto a reforzar si el sistema fuera a recibir escrituras concurrentes desde varias cajas al mismo tiempo. Tampoco hay borrado ni reemplazo de claves con actualización atómica entre el treap y SQLite dentro de una transacción conjunta, así que ante una caída del proceso exactamente entre la escritura en SQLite y la replicación en el treap, el estado en memoria podría quedar un paso desactualizado hasta el próximo reinicio, momento en el que CargarProductos vuelve a reconstruir el treap desde SQLite y corrige cualquier desfase.

En conjunto, el treap le da a este sistema justo la combinación que necesita un punto de venta simple: inserciones que pueden llegar en cualquier orden sin degradar el rendimiento, búsquedas por código prácticamente instantáneas, y un catálogo siempre disponible en orden alfabético sin costo adicional, todo con una implementación notablemente más simple de escribir y mantener que la de un AVL o un árbol Rojo-Negro equivalente.
