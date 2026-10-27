# Ejemplo compose

App de prueba de Kalos Tech para los tenants con varios contenedores: un servidor en Go (`web`), Postgres (`db`) y Redis (`cache`). Ver `pantheon/docs/features/tenants-compose.md`.

```sh
docker compose up --build   # http://localhost:8080
```
