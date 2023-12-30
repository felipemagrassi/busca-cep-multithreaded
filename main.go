package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

type ViaCepResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type Address struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
	Service      string `json:"service"`
}

func main() {
	r := chi.NewRouter()
	r.Use(middleware.Logger)

	r.Get("/", BuscaCepHandler)

	http.ListenAndServe(":8081", r)
}

func BuscaCepHandler(w http.ResponseWriter, r *http.Request) {
	onlyNumbersRegex := regexp.MustCompile(`[^\d]`)
	cepParam := r.URL.Query().Get("cep")
	cep := onlyNumbersRegex.ReplaceAllString(cepParam, "")

	ch := make(chan *Address)

	if cep == "" {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cep não informado"))
		return
	}

	if len(cep) != 8 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Cep inválido, deve conter 8 caracteres"))
		return
	}
	fmt.Println("Requesting address for cep:", cep)

	go ViaCEPHandler(cep, ch)
	go BrasilAPIHandler(cep, ch)

	select {
	case address := <-ch:
		w.WriteHeader(http.StatusOK)
		w.Header().Set("Content-Type", "application/json")

		if address.Street == "" {
			fmt.Println("Endereço não encontrado no serviço: ", address.Service)
		} else {
			fmt.Println("Cep:", address.Cep)
			fmt.Println("State:", address.State)
			fmt.Println("City:", address.City)
			fmt.Println("Neighborhood:", address.Neighborhood)
			fmt.Println("Street:", address.Street)
			fmt.Println("Service:", address.Service)
		}

		json.NewEncoder(w).Encode(address)
	case <-time.After(1 * time.Second):
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Tempo limite excedido"))
	}
}

func ViaCEPHandler(cep string, ch chan *Address) (*Address, error) {
	req, err := http.Get("https://viacep.com.br/ws/" + cep + "/json/")
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var response ViaCepResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	address := &Address{
		Cep:          response.Cep,
		State:        response.Uf,
		City:         response.Localidade,
		Neighborhood: response.Bairro,
		Street:       response.Logradouro,
		Service:      "Viacep",
	}

	ch <- address

	return address, nil
}

func BrasilAPIHandler(cep string, ch chan *Address) (*Address, error) {
	req, err := http.Get("https://brasilapi.com.br/api/cep/v1/" + cep)
	if err != nil {
		return nil, err
	}

	defer req.Body.Close()

	body, err := io.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}

	var response BrasilAPIResponse
	err = json.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	address := &Address{
		Cep:          response.Cep,
		State:        response.State,
		City:         response.City,
		Neighborhood: response.Neighborhood,
		Street:       response.Street,
		Service:      "BrasilAPI",
	}

	ch <- address

	return address, nil
}
