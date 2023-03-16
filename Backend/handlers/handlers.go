// O pacote handlers faz a ligação entre as requisições http e suas respectivas funções
// ligando o Crud para manipulação do banco de dados, ou chamando diretamente as funções
// de ordenação no DataManager
// Handlers também realiza o parsing entre JSON e Objeto
package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/Bernardo46-2/AEDS-III/crud"
	"github.com/Bernardo46-2/AEDS-III/dataManager"
	"github.com/Bernardo46-2/AEDS-III/logger"
	"github.com/Bernardo46-2/AEDS-III/models"
	"github.com/Bernardo46-2/AEDS-III/utils"
)

// GetPagesNumber retorna a quantidade de paginas disponiveis
func GetPagesNumber(w http.ResponseWriter, r *http.Request) {
	// Recuperar ID e ler arquivo
	numeroPaginas, err := crud.ReadPagesNumber()

	// Resposta
	if err != nil {
		writeError(w, http.StatusInternalServerError, 2)
		return
	}

	writeJson(w, numeroPaginas)
}

// GetAllPokemon recupera os 60 pokemons a partir do ID fornecido
func GetAllPokemon(w http.ResponseWriter, r *http.Request) {
	// Recuperar ID e ler arquivo
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pokemon, err := crud.ReadAll(page)

	// Resposta
	if err != nil {
		writeError(w, http.StatusInternalServerError, 2)
		return
	}

	writeJson(w, pokemon)
	logger.Println("GET", "Id de numero "+strconv.Itoa(int(pokemon[0].Numero))+" ate "+strconv.Itoa(int(pokemon[len(pokemon)-1].Numero)))
}

// GetPokemon recupera o pokemon pelo ID fornecido
func GetPokemon(w http.ResponseWriter, r *http.Request) {
	// recuperar ID e ler do arquivo
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	pokemon, err := crud.Read(id)

	// Gera resposta de acordo com o resultado
	if err != nil {
		writeError(w, http.StatusInternalServerError, 2)
		return
	}

	writeJson(w, pokemon)
	logger.Println("GET", "Id de numero "+strconv.Itoa(id))
}

// PostPokemon adiciona o pokemon ao banco de dados
func PostPokemon(w http.ResponseWriter, r *http.Request) {

	// Desserialização
	var pokemon models.Pokemon
	err := json.NewDecoder(r.Body).Decode(&pokemon)
	defer r.Body.Close()
	if err != nil {
		writeError(w, http.StatusBadRequest)
		return
	}

	// Create
	id, err := crud.Create(pokemon)

	// Resposta
	if err != nil {
		writeError(w, http.StatusInternalServerError, 3)
		return
	}
	pokemonID := models.PokemonID{ID: id}
	writeJson(w, pokemonID)
	logger.Println("POST", "Id de numero "+strconv.Itoa(id)+" adicionado")
}

// PutPokemon recebe um json e atualiza o valor no banco de dados
// de acordo com o dado recebido
func PutPokemon(w http.ResponseWriter, r *http.Request) {
	//  Desserialização
	var pokemon models.Pokemon
	err := json.NewDecoder(r.Body).Decode(&pokemon)
	defer r.Body.Close()

	if err != nil {
		writeError(w, http.StatusBadRequest)
		return
	}

	// Update
	err = crud.Update(pokemon)

	// Resposta
	if err != nil {
		writeError(w, http.StatusInternalServerError)
		return
	}

	writeSuccess(w, 4)
	logger.Println("PUT", "Id de numero "+strconv.Itoa(int(pokemon.Numero))+" atualizado")
}

// DeletePokemon recebe um ID, pesquisa no banco de dados
// e se existir efetiva sua remoção logica
func DeletePokemon(w http.ResponseWriter, r *http.Request) {
	// Recupera id
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))

	// Delete
	_, err := crud.Delete(id)

	// Resposta
	if err != nil {
		writeError(w, http.StatusInternalServerError, 5)
		return
	}

	writeSuccess(w, 5)
	logger.Println("DELETE", "Id de numero "+strconv.Itoa(id)+" deletado")
}

// LoadDatabase faz o carregamento do arquivo CSV e o serializa em binario
func LoadDatabase(w http.ResponseWriter, r *http.Request) {
	// Import
	dataManager.ImportCSV().CsvToBin()

	// Resposta
	writeSuccess(w, 6)
	logger.Println("INFO", "Database Recarregada")
}

// ToKatakana recebe uma string em alfabeto romato, converte para
// o padrão katakana da linguagem japonesa e retorna a string
// convertida
func ToKatakana(w http.ResponseWriter, r *http.Request) {
	// Intercepta
	stringToConvert := r.URL.Query().Get("stringToConvert")

	// Converte
	convertedString := utils.ToKatakana(stringToConvert)

	// Resposta
	writeJson(w, convertedString)
}

// IntercalacaoComum realiza a ordenação externa do BD através da intercalação balanceada.
//
// A função lê buffers do arquivo, ordena internamente, salva em diferentes arquivos e,
// por fim, os une utilizando mergesort.
func IntercalacaoComum(w http.ResponseWriter, r *http.Request) {
	// Ordena
	dataManager.IntercalacaoBalanceadaComum()

	// Resposta
	writeSuccess(w, 7)
	logger.Println("INFO", "Database Ordenada (Intercalacao Comum)")
}

// IntercalacaoVariavel realiza a ordenação externa do BD através da intercalação variavel
//
// A função lê buffers do arquivo e salva em novos arquivos enquanto estiver ordenado
// Cria um novo arquivo para cada buffer desalinhado e, por fim, os une utilizando
// mergesort externo.
func IntercalacaoVariavel(w http.ResponseWriter, r *http.Request) {
	// Ordena
	dataManager.IntercalacaoBalanceadaVariavel()

	// Resposta
	writeSuccess(w, 8)
	logger.Println("INFO", "Database Ordenada (Intercalacao Comum)")
}

// SelecaoPorSubstituicao realiza a ordenação externa do BD através de um heap minimo
//
// A função lê buffers do arquivo e insere em um heap minimo de tamanho fixo,
// a cada inserção o heap é desmontado em arquivos temporarios e inserido novos registros.
// Por fim os arquivos sao unidos em mergesort externo
func SelecaoPorSubstituicao(w http.ResponseWriter, r *http.Request) {
	// Ordena
	dataManager.IntercalacaoPorSubstituicao()

	// Resposta
	writeSuccess(w, 9)
	logger.Println("INFO", "Database Ordenada (Intercalacao Comum)")
}

// writeError recebe um erro de http responde e um id de erro interno,
// faz o parsing do modelo e gera uma resposta em formato json com o erro fornecido
func writeError(w http.ResponseWriter, codes ...int) {
	// Preparacao da resposta http
	w.Header().Set("Content-Type", "application/json")
	code := codes[0]
	w.WriteHeader(code)
	if len(codes) > 1 {
		code = codes[1]
	}

	// Gera uma resposta json personalizada
	json.NewEncoder(w).Encode(models.ErrorResponse(code))
}

// writeSuccess gera uma resposta http de sucesso (200) e
// faz o parsing do modelo de sucesso para uma resposta json com a mensagem
// da ação realizada
func writeSuccess(w http.ResponseWriter, code int) {
	// Preparação da resposta http
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Gera uma resposta json personalizada
	json.NewEncoder(w).Encode(models.SuccessResponse(code))
}

// writeJson recebe qualquer tipo de dado ou struct e serializa o dado
// em formato json, gerando junto uma resposta de sucesso ou erro
func writeJson(w http.ResponseWriter, v any) {
	// Serialização
	jsonData, err := json.Marshal(v)

	// Resposta
	if err != nil {
		writeError(w, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}
