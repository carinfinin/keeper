# Keeper хранилище ключей, паролей и бинарных данных

## Для запуска сервера неоюходимо:

- Сгенерировать ключи командой 
```bash
make gen_private_key
make gen_public_key
```
- Запустить миграции
```bash
make migrate
```
- (сбилдить) Запустить сервер
```bash
make build_server
make run_server
```

## Для запуска клиена неоюходимо:
- сбилдить клиент
```bash
make build_client
```
- сбилдить воркер
```bash
make build_worker
``````
- запусть установку крон воркера
```bash
make setup_worker
```