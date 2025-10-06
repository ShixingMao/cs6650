# Product API (Online Store)

### Run locally
```bash
go run main.go

```
### Test locally with curl commands
```post
curl -X POST http://localhost:8080/products/1/details \
-H "Content-Type: application/json" \
-d '{"product_id":1,"sku":"SKU123","manufacturer":"Acme","category_id":2,"weight":100,"some_other_id":3}'
```
```get
curl http://localhost:8080/products/1
```
### Deploy
cd terraform
terraform init
terraform apply
