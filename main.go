// Ejemplo compose: app de prueba de Kalos Tech para los tenants con docker compose
// (pantheon/docs/features/tenants-compose.md). Cuenta visitas en Postgres y le pregunta a Redis si
// está vivo, para comprobar que los servicios se encuentran por nombre (db, cache).
package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func redisPing(addr string) string {
	c, err := net.DialTimeout("tcp", addr, 2*time.Second)
	if err != nil {
		return "error: " + err.Error()
	}
	defer c.Close()
	c.SetDeadline(time.Now().Add(2 * time.Second))
	fmt.Fprint(c, "PING\r\n")
	line, err := bufio.NewReader(c).ReadString('\n')
	if err != nil {
		return "error: " + err.Error()
	}
	return line[:len(line)-2]
}

func main() {
	ctx := context.Background()
	db, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	// Postgres puede tardar en partir: la app reintenta (en kosmos no hay depends_on).
	for {
		if _, err = db.Exec(ctx, "CREATE TABLE IF NOT EXISTS visitas (n bigint NOT NULL)"); err == nil {
			break
		}
		log.Printf("esperando a la base de datos: %v", err)
		time.Sleep(2 * time.Second)
	}
	db.Exec(ctx, "INSERT INTO visitas SELECT 0 WHERE NOT EXISTS (SELECT 1 FROM visitas)")

	http.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) { fmt.Fprintln(w, "ok") })
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		var n int64
		if err := db.QueryRow(r.Context(), "UPDATE visitas SET n = n + 1 RETURNING n").Scan(&n); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		fmt.Fprintf(w, "Ejemplo compose: visita %d (Postgres) · Redis: %s\n", n, redisPing(os.Getenv("REDIS_ADDR")))
	})
	log.Println("escuchando en :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
