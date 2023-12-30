# Serviço para buscar um CEP e retornar um endereço completo

Esse serviço busca um CEP em duas APIs (ViaCEP e BrasilAPI) e retorna um endereço completo com os dados encontrados em uma das APIs.


## Retorno 

```go
type Address struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}
```

O retorno será um JSON com os dados encontrados em uma das APIs. Caso não seja encontrado o CEP, o retorno será um JSON vazio.

O campo `service` indica de qual API foi encontrado o CEP. Pode ser `Viacep` ou `BrasilAPI`.
