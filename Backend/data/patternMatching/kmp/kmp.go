package kmp

import (
	"strings"

	"github.com/Bernardo46-2/AEDS-III/data/binManager"
	"github.com/Bernardo46-2/AEDS-III/data/indexes/invertedIndex"
)

const (
	// PatternSize é o tamanho máximo do padrão a ser procurado. É usado para definir o tamanho do array no pré-processamento KMP.
	PatternSize int = 100
)

// SearchNext busca a última ocorrência da string de busca (needle) na string de destino (haystack).
//
// Parâmetros:
//
//	haystack: A string de destino onde a busca será realizada.
//	needle: A string de busca que será procurada na string de destino.
//
// Retorno:
//
//	A posição da última ocorrência da string de busca na string de destino. Retorna -1 se a string de busca não for encontrada.
//
// A função utiliza o algoritmo KMP para realizar a busca. Se a string de busca for encontrada, retorna a posição da última ocorrência.
// Se não for encontrada, retorna -1.
func SearchNext(haystack string, needle string) int {
	retSlice := kmp(haystack, needle)
	if len(retSlice) > 0 {
		return retSlice[len(retSlice)-1]
	}

	return -1
}

// SearchString realiza a busca da string de busca (needle) na string de destino (haystack), ignorando a diferença entre letras maiúsculas e minúsculas.
//
// Parâmetros:
//
//	haystack: A string de destino onde a busca será realizada.
//	needle: A string de busca que será procurada na string de destino.
//
// Retorno:
//
//	Um slice de inteiros contendo as posições de todas as ocorrências da string de busca na string de destino. Retorna um slice vazio se a string de busca não for encontrada.
//
// A função utiliza o algoritmo KMP para realizar a busca. Ambas as strings de busca e de destino são convertidas para letras minúsculas antes da busca, para garantir que a busca seja insensível a maiúsculas e minúsculas.
func SearchString(haystack string, needle string) []int {
	return kmp(strings.ToLower(haystack), strings.ToLower(needle))
}

// kmp realiza a busca do algoritmo Knuth-Morris-Pratt (KMP).
//
// Parâmetros:
//
//	haystack: A string principal onde a busca será realizada.
//	needle: A string de busca.
//
// Retorno:
//
//	Um slice de inteiros que contém todas as posições iniciais na string haystack onde a string needle é encontrada.
//
// A função primeiro cria a tabela de prefixos para a string needle.
// Em seguida, passa pelos caracteres da string haystack.
// Se um caractere na posição atual na string haystack é igual ao caractere na posição atual na string needle,
// avança na string haystack e needle.
//
// Se os caracteres não são iguais, move a posição atual na string needle para a próxima posição na tabela de prefixos que foi criada anteriormente.
// Se encontrou uma correspondência para a string needle na string haystack, armazena a posição inicial na string haystack na lista de resultados e
// move a posição atual na string needle para a próxima posição na tabela de prefixos.
func kmp(haystack string, needle string) []int {
	next := preKMP(needle)
	i := 0
	j := 0
	m := len(needle)
	n := len(haystack)

	x := []byte(needle)
	y := []byte(haystack)
	var ret []int

	// se algum dos valores sao nulos
	if m == 0 || n == 0 {
		return ret
	}

	// se string for maior do que texto
	if n < m {
		return ret
	}

	// Percorre os caracteres na string haystack
	for j < n {
		// Se o caractere atual na haystack não for igual ao da needle, atualiza i para o próximo valor na tabela de prefixos
		for i > -1 && x[i] != y[j] {
			i = next[i]
		}
		i++
		j++

		// Se i é maior ou igual ao comprimento da needle, encontramos uma correspondência e a adicionamos ao slice de retorno
		if i >= m {
			ret = append(ret, j-i)
			i = next[i]
		}
	}

	return ret
}

// preKMP realiza o pré-processamento da string de busca e retorna um array que contém
// a maior borda própria de cada prefixo da string de busca. Esta tabela será usada
// pelo algoritmo KMP para pular as comparações de caracteres que já foram comparados.
//
// Parâmetros:
//
//	x: A string de busca.
//
// Retorno:
//
//	Um array de tamanho PatternSize que contém a maior borda própria de cada prefixo
//	da string de busca.
//
// A função inicializa i e j com 0 e -1 respectivamente e define o primeiro valor de
// kmpNext como -1. A variável i é o índice para percorrer os caracteres na string de busca,
// enquanto j mantém a maior borda própria do prefixo atual.
//
// A função então entra em um loop, onde para cada caractere na string de busca, se o caractere
// não for igual ao caractere na posição j, atualiza j para o valor de kmpNext na posição j.
//
// Se o caractere for igual, incrementa i e j, e se o próximo caractere também for igual,
// define o valor de kmpNext na posição i como o valor de kmpNext na posição j. Caso contrário,
// define o valor de kmpNext na posição i como j.
func preKMP(x string) [PatternSize]int {
	var i, j int
	length := len(x) - 1
	var kmpNext [PatternSize]int
	i = 0
	j = -1
	kmpNext[0] = -1 // A borda própria mais longa de uma string vazia é -1

	// Percorre a string de busca
	for i < length {
		// Encontra a maior borda própria do prefixo atual
		for j > -1 && x[i] != x[j] {
			j = kmpNext[j]
		}

		i++
		j++

		// Se o próximo caractere também for igual, define o valor de kmpNext na posição i como o valor de kmpNext na posição j
		// Caso contrário, define o valor de kmpNext na posição i como j
		if x[i] == x[j] {
			kmpNext[i] = kmpNext[j]
		} else {
			kmpNext[i] = j
		}
	}
	return kmpNext
}

// SearchPokemon realiza uma busca por um termo específico (search) em um campo específico
// (field) dos registros de Pokemon, retornando os documentos que contêm o Id do pokemon
// que possui aquele termo e a frequencia de aparição.
//
// Param:
//
//	search: O termo de busca que será procurado no campo especificado dos registros de Pokemon.
//	field: O campo dos registros de Pokemon onde o termo de busca será procurado.
//
// Retorno:
//
//	Um slice de ScoredDocument, onde cada ScoredDocument contém o ID do documento e a pontuação,
//	que é o número de ocorrências do termo de busca no campo especificado. Retorna um slice vazio
//	se o termo de busca não for encontrado.
//
// A função inicia o controlador de leitura e lê todos os registros de Pokemon um por um. Para
// cada registro de Pokemon, a função realiza a busca pelo termo de busca no campo especificado
// usando a função SearchString. Se o termo de busca for encontrado, um ScoredDocument é criado
// com o ID do documento sendo o número do Pokemon e a pontuação sendo o número de ocorrências
// do termo de busca. O ScoredDocument é então adicionado ao slice de ScoredDocument.
func SearchPokemon(search string, field string) (scoredDocuments []invertedIndex.ScoredDocument) {
	// Abertura do controlador de leitura
	controller, _ := binManager.InicializarControleLeitura(binManager.BIN_FILE)
	scoredDocuments = make([]invertedIndex.ScoredDocument, 0)

	// Ler enquanto nao acontecer FEOF
	for err := controller.ReadNext(); err == nil; err = controller.ReadNext() {
		// Se nao possuir lapide pesquisar
		if !controller.RegistroAtual.IsDead() {
			needle := SearchString(controller.RegistroAtual.Pokemon.GetField(field), search)
			if len(needle) > 0 {
				scoredDocuments = append(scoredDocuments, invertedIndex.ScoredDocument{DocumentID: int64(controller.RegistroAtual.Pokemon.Numero), Score: len(needle)})
			}
		}
	}

	return scoredDocuments
}