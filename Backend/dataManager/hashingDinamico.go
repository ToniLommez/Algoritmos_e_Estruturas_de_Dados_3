// Dynamic Hash - Fornece uma implementação de hash dinâmica.
//
// Belo Horizonte - 28/03/2029
// Marcos Antonio Lommez Candido Ribeiro
// Bernardo Marques Fernandes
//
// Este pacote oferece uma implementação de hash dinâmica que permite a criação
// e gerenciamento de hashes que se adaptam de acordo com o tamanho do conjunto de dados.
// A hash dinâmica é útil em cenários onde o tamanho do conjunto de dados pode variar
// significativamente e onde o desempenho de pesquisa é crítico.
//
// Para começar, adapte a funcao StartHashFile() e recordToBucketRecord(registro Registro)
// para receberem o tipo de registro que voce deseja criar a hash.
// Para alem disso utilize uma função de leitura de arquivo para fazer o import na funcao
// StartHashFile()
//
// Consulte a documentação das funções e tipos individuais para obter exemplos de uso e
// informações adicionais.

package dataManager

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/Bernardo46-2/AEDS-III/models"
	"github.com/Bernardo46-2/AEDS-III/utils"
)

const BUCKETS_FILE string = "data/Hash_Buckets.bin"
const DIRECTORY_FILE string = "data/Hash_Directory.bin"

// ====================================== Structs ====================================== //

// DinamicHash representa uma tabela de hash dinâmica.
// A tabela de hash usa uma estrutura de diretório para manter
// ponteiros para os buckets armazenados em um arquivo.
type DinamicHash struct {
	bucketFile    *os.File  // O arquivo que armazena os buckets.
	directoryFile *os.File  // O arquivo que armazena o diretório.
	loadFactor    int64     // Fator de carga máximo de um bucket.
	bucketSize    int64     // Tamanho em bytes de cada bucket.
	directory     Directory // O diretório que mantém os ponteiros para os buckets.
}

// Directory representa o diretório da tabela de hash dinâmica.
// Ele mantém ponteiros para os buckets e o nível de profundidade atual (p).
//
// 'p' pode ser lido tanto quanto power quanto como profundidade.
type Directory struct {
	p             int64   // Nível de profundidade atual.
	bucketPointer []int64 // Ponteiros para os buckets no arquivo bucketFile.
	garbage       []int64 // Lista de buckets vazios para reaproveitar
}

// Bucket representa um bucket na tabela de hash dinâmica.
// Cada bucket tem uma capacidade máxima definida pelo loadFactor.
// Ele armazena registros do tipo BucketRecord.
type Bucket struct {
	ActualPower int64          // Profundidade local do bucket.
	CurrentSize int64          // Número atual de registros no bucket.
	Records     []BucketRecord // Coleção de registros no bucket.
}

// BucketRecord representa um registro em um bucket na tabela de hash dinâmica.
// Cada registro contém um ID e um endereço.
type BucketRecord struct {
	ID      int64 // ID do registro.
	Address int64 // Endereço original do registro.
}

// =================================== Dinamic Hash ==================================== //

// newHash inicializa uma nova estrutura de hash dinamico nova
// pronta para ser preenchida
func newHash(bucketPath string, directoryPath string, size int64) DinamicHash {
	// Inicializa arquivo de hashing
	bucketFile, _ := os.Create(bucketPath)
	directoryFile, _ := os.Create(directoryPath)

	// Definições das structs
	d := Directory{
		p:             1,
		bucketPointer: make([]int64, 2),
	}
	hash := DinamicHash{
		directory:     d,
		loadFactor:    size + 1, // para suportar explosao
		bucketFile:    bucketFile,
		directoryFile: directoryFile,
	}

	// Calculando espaço a ser gasto por bucket
	ActualPowerSize := int64(binary.Size(Bucket{}.ActualPower))
	currentSizeSize := int64(binary.Size(Bucket{}.CurrentSize))
	bucketRecordSize := int64(binary.Size(BucketRecord{}))
	hash.bucketSize = ActualPowerSize + currentSizeSize + (hash.loadFactor * bucketRecordSize)

	// Preenchendo arquivos com templates vazios
	bucketAddress := hash.initializeNewBucket(hash.getBucketCount())

	// Guardando enderecos dos buckets criados
	for i := 0; i < hash.getBucketCount(); i++ {
		hash.directory.bucketPointer[i] = bucketAddress[i]
	}

	return hash
}

// loadDirectory carrega o diretorio da hash para a memoria primaria
func loadDinamicHash(directoryPath string) (hash DinamicHash, err error) {
	directoryFile, err := os.OpenFile(directoryPath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		StartHashFile()
		err = fmt.Errorf("arquivo hash inexistente, fazendo upload dos registros e criando hash nova")
	}

	var ptr int

	// Dados da hash dinamica
	buffer, _ := io.ReadAll(directoryFile)
	bucketPath, ptr := utils.BytesToString(buffer, ptr)
	loadFactor, ptr := utils.BytesToInt64(buffer, ptr)
	bucketSize, ptr := utils.BytesToInt64(buffer, ptr)

	// Dados do bucket
	p, ptr := utils.BytesToInt64(buffer, ptr)
	bucketPointerLen, ptr := utils.BytesToInt64(buffer, ptr)
	bucketPointer := make([]int64, bucketPointerLen)
	for i := 0; i < int(bucketPointerLen); i++ {
		bucketPointer[i], ptr = utils.BytesToInt64(buffer, ptr)
	}
	bucketGarbageLen, ptr := utils.BytesToInt64(buffer, ptr)
	bucketGarbage := make([]int64, bucketGarbageLen)
	for i := 0; i < int(bucketGarbageLen); i++ {
		bucketGarbage[i], ptr = utils.BytesToInt64(buffer, ptr)
	}

	// Inicializando as structs
	bucketFile, _ := os.OpenFile(bucketPath, os.O_RDWR|os.O_CREATE, 0644)

	directory := Directory{
		p:             p,
		bucketPointer: bucketPointer,
		garbage:       bucketGarbage,
	}
	hash = DinamicHash{
		directory:     directory,
		loadFactor:    loadFactor,
		bucketFile:    bucketFile,
		directoryFile: directoryFile,
		bucketSize:    bucketSize,
	}

	return hash, err
}

// Close salva o arquivo do diretorio com os dados atuais
// e em seguida fecha os arquivos dependentes abertos
func (hash *DinamicHash) Close() {
	hash.directoryFile.Seek(0, io.SeekStart)

	// Dados da hash dinamica
	binary.Write(hash.directoryFile, binary.LittleEndian, int32(len(hash.bucketFile.Name())))
	binary.Write(hash.directoryFile, binary.LittleEndian, []byte(hash.bucketFile.Name()))
	binary.Write(hash.directoryFile, binary.LittleEndian, hash.loadFactor)
	binary.Write(hash.directoryFile, binary.LittleEndian, hash.bucketSize)

	// Dados do bucket
	binary.Write(hash.directoryFile, binary.LittleEndian, hash.directory.p)
	binary.Write(hash.directoryFile, binary.LittleEndian, int64(len(hash.directory.bucketPointer)))
	for i := 0; i < len(hash.directory.bucketPointer); i++ {
		binary.Write(hash.directoryFile, binary.LittleEndian, hash.directory.bucketPointer[i])
	}
	binary.Write(hash.directoryFile, binary.LittleEndian, int64(len(hash.directory.garbage)))
	for i := 0; i < len(hash.directory.garbage); i++ {
		binary.Write(hash.directoryFile, binary.LittleEndian, hash.directory.garbage[i])
	}

	if err := hash.bucketFile.Close(); err != nil {
		fmt.Printf("Erro ao fechar/salvar bucket")
	}
	if err := hash.directoryFile.Close(); err != nil {
		fmt.Printf("Erro ao fechar/salvar diretorio")
	}
}

// increasePower aumenta o 'power' do diretorio e cria
// os novos ponteiros para os buckets existentes
func (hash *DinamicHash) increasePower() {
	hash.directory.p++
	novoTamanhoBucket := 1 << hash.directory.p
	novoBucket := make([]int64, novoTamanhoBucket)
	copy(novoBucket, hash.directory.bucketPointer)
	hash.directory.bucketPointer = novoBucket

	for i, j := (hash.getBucketCount() >> 1), 0; i < hash.getBucketCount(); i, j = i+1, j+1 {
		hash.directory.bucketPointer[i] = hash.directory.bucketPointer[j]
	}
}

// PrintHash é uma funcao pertencente a struct DinamicHash
// que permite fazer o debug da hash.
//
// Chame-a a cada inserção para um debug completo, ou ao final
// para uma apresentação da hash atual
//
// Atualmente a função esta formatada para hash de tamanho 8, mas funciona
// com outros tamanhos, a implementação generica ficara para um futuro
func (hash *DinamicHash) PrintHash() {
	// Mapeamento de variaveis printadas, para nao printar indices repetidos para um mesmo bucket
	seen := make(map[int64]bool)

	fmt.Println()
	fmt.Printf("||-----------------------------------------------------------------------------------------------------------------------------------------||\n")
	fmt.Printf("||  Num  |  Offs  |  Power  |  Count  ||                                           Key / Address                                           ||\n")
	fmt.Printf("||-------|--------|---------|---------||---------------------------------------------------------------------------------------------------||\n")

	for i := 0; i < len(hash.directory.bucketPointer); i++ {
		// Printa apenas buckets nao repetidos
		if !seen[hash.directory.bucketPointer[i]] {
			// Recupera a posição a partir do diretorio, da parsing e printa
			bucket := hash.readBucket(int64(i))

			// Itera sobre os valores e printa
			fmt.Printf("|| [%3d] | %5x  |    %d    |    %d    ||  ", i, hash.directory.bucketPointer[i], bucket.ActualPower, bucket.CurrentSize)
			for i := 0; i < len(bucket.Records)-1; i++ {
				if bucket.Records[i].ID != 0 {
					fmt.Printf("{%3d %5x} ", bucket.Records[i].ID, bucket.Records[i].Address)
				} else {
					fmt.Printf("            ")
				}
			}
			fmt.Printf(" ||\n")
			seen[hash.directory.bucketPointer[i]] = true
		} else {
			fmt.Printf("|| [%3d] | %5x  |         |         ||                                                                                                   ||\n", i, hash.directory.bucketPointer[i])
		}
	}
	fmt.Printf("||-----------------------------------------------------------------------------------------------------------------------------------------||\n")
	fmt.Printf("\n")
}

// StartHashFile cria um arquivo de hash para a pokedex e por
// fim printa o conteudo da hash
func StartHashFile() {
	// Inicializando controle e hash vazia
	c, err := inicializarControleLeitura(BIN_FILE)
	hash := newHash(BUCKETS_FILE, DIRECTORY_FILE, 8)

	// Parsing e inclusao na hash, se acabar o arquivo sera retornado um erro io.EOF
	for i := 0; i < int(c.TotalRegistros) && err == nil; i++ {
		err = c.ReadNext()
		if c.RegistroAtual.Lapide != 1 {
			r := recordToBucketRecord(*c.RegistroAtual)
			hash.addRecord(r)
		}
	}

	// Fechando hash e salvando diretorio
	hash.Close()
}

// addRecord adiciona um BucketRecord a estrutura de hash.
// A função utiliza as variaveis nativas da estrutura DinamicHash para
// recuperar o arquivo e seus metadados
//
// Caso um bucket de "localPower" == "hashPower" a hash sera aumentada e um novo bucket criado.
// Caso o "localPower" < "hashPower" um novo bucket sera criado.
// Por fim se nao estourar apenas insere
func (hash *DinamicHash) addRecord(r BucketRecord) {
	// Recuperar e dar parsing no bucket a ser editado
	pos := int64(r.ID) % int64(hash.getBucketCount())
	bucket := hash.readBucket(pos)

	// Atualiza o bucket com o novo valor
	if bucket.CurrentSize == hash.loadFactor-1 {
		// Se o bucket tiver apenas 1 ponteiro aumentar p em +1, se nao so atualiza o bucket
		if bucket.ActualPower == hash.directory.p {
			hash.increasePower()
		}

		// Criação do novo bucket
		address := hash.initializeNewBucket(1)
		newPos := bucket.getBucketPower() + pos
		if newPos >= int64(hash.getBucketCount()) {
			newPos = pos
			pos %= bucket.getBucketPower()
		}
		hash.directory.bucketPointer[newPos] = address[0]
		bucket.ActualPower++

		// Limpeza e reinsercao
		bucket1 := newBucket(bucket.ActualPower, hash.loadFactor)
		bucket2 := newBucket(bucket.ActualPower, hash.loadFactor)
		bucket.Records[bucket.CurrentSize] = r
		for i, b1, b2 := 0, 0, 0; i < len(bucket.Records); i++ {
			if bucket.Records[i].ID%bucket.getBucketPower() == pos {
				bucket1.Records[b1] = bucket.Records[i]
				bucket1.CurrentSize++
				b1++
			} else {
				bucket2.Records[b2] = bucket.Records[i]
				bucket2.CurrentSize++
				b2++
			}
		}

		// Gravando bucket atual e novo em arquivo
		hash.insertIntoBucket(pos, bucket1.ActualPower, bucket1.CurrentSize, bucket1.Records)
		hash.insertIntoBucket(newPos, bucket2.ActualPower, bucket2.CurrentSize, bucket2.Records)
	} else {
		// Apenas insere ao final
		bucket.Records[bucket.CurrentSize] = r
		hash.insertIntoBucket(pos, bucket.ActualPower, bucket.CurrentSize+1, bucket.Records)
	}
}

// ====================================== Bucket ======================================= //

// newBucket retorna um bucket preparado para ser preenchido
func newBucket(actualPower int64, loadFactor int64) Bucket {
	return Bucket{
		ActualPower: actualPower,
		CurrentSize: 0,
		Records:     make([]BucketRecord, loadFactor),
	}
}

// initializeNewBucket inicializa a quantidade de buckets fornecidos no arquivo
// e retorna um array de enderecos dos buckets inicializados
//
// Caso exista alguma posicao marcada como garbage collector, esta sera retornada
// para nao precisar criar um novo espaço
func (hash *DinamicHash) initializeNewBucket(numberOfBuckets int) []int64 {
	bucketAddress := make([]int64, numberOfBuckets)
	hash.bucketFile.Seek(0, io.SeekEnd)

	for i := 0; i < numberOfBuckets; i++ {
		if len(hash.directory.garbage) > 0 {
			// Reutilizando uma posição do garbage
			bucketAddress[i] = hash.directory.garbage[0]
			hash.directory.garbage = hash.directory.garbage[1:]

			// Movendo o cursor do arquivo para a posição reutilizada
			hash.bucketFile.Seek(bucketAddress[i], io.SeekStart)
		} else {
			// Criando um novo espaço no arquivo
			bucketAddress[i], _ = hash.bucketFile.Seek(0, io.SeekEnd)
		}

		// Escrevendo os dados no bucket
		binary.Write(hash.bucketFile, binary.LittleEndian, hash.directory.p)                      // ActualPower
		binary.Write(hash.bucketFile, binary.LittleEndian, int64(0))                              // CurrentSize
		binary.Write(hash.bucketFile, binary.LittleEndian, make([]BucketRecord, hash.loadFactor)) // Records
	}

	return bucketAddress
}

// readBucket recebe a posição do bucket na hash, realiza o parsing
// e retorna o bucket formatado
func (hash *DinamicHash) readBucket(pos int64) Bucket {

	// Parsing dos metadados
	var ID int64
	var Address int64
	var ptr int
	var bucket Bucket

	if pos < 0 {
		return bucket
	}

	// Recuperando a posição do arquivo na hash e lendo os dados cruamente
	hash.bucketFile.Seek(hash.directory.bucketPointer[pos], io.SeekStart)
	data := make([]byte, hash.bucketSize)
	hash.bucketFile.Read(data)

	bucket.Records = make([]BucketRecord, hash.loadFactor)
	bucket.ActualPower, ptr = utils.BytesToInt64(data, ptr)
	bucket.CurrentSize, ptr = utils.BytesToInt64(data, ptr)

	// Parsing dos registros
	for i := int64(0); i < hash.loadFactor; i++ {
		ID, ptr = utils.BytesToInt64(data, ptr)
		Address, ptr = utils.BytesToInt64(data, ptr)
		bucket.Records[i] = BucketRecord{
			ID:      ID,
			Address: Address,
		}
	}

	return bucket
}

// getBucketCount retorna a quantidade de caminhos existentes na função hash.
//
// A função calcula 2^p
func (hash *DinamicHash) getBucketCount() int {
	return 1 << hash.directory.p
}

// getBucketPower retorna a potencia de 2 do numero fornecido.
//
// A funcao é usada para saber a potencia atual do bucket
func (b *Bucket) getBucketPower() int64 {
	return 1 << b.ActualPower
}

// insertIntoBucket insere um bucket em arquivo,
// onde 'pos' é a posicao na hash, o 'power' local, o preenchimento atual
// e por fim os records
//
// A função insertIntoBucket evita a criação de uma struct Bucket
// pois é possivel fazer a gravação diretamente no arquivo
// e assim economizando espaço.
func (hash *DinamicHash) insertIntoBucket(pos int64, power int64, currentSize int64, records []BucketRecord) {
	// Recuperando a posição do bucket no arquivo e escrevendo
	hash.bucketFile.Seek(hash.directory.bucketPointer[pos], io.SeekStart)
	binary.Write(hash.bucketFile, binary.LittleEndian, power)
	binary.Write(hash.bucketFile, binary.LittleEndian, currentSize)
	binary.Write(hash.bucketFile, binary.LittleEndian, records)
}

// recordToBucketRecord transforma um registro de leitura de arquivo
// em um registro de bucket
func recordToBucketRecord(registro Registro) BucketRecord {
	return BucketRecord{
		ID:      int64(registro.Pokemon.Numero),
		Address: registro.Endereco,
	}
}

// pokemonToBucketRecord retorna uma struct de BucketRecord inicializada
// com os valores fornecidos
func pokemonToBucketRecord(pokemon models.Pokemon, address int64) BucketRecord {
	return BucketRecord{
		ID:      int64(pokemon.Numero),
		Address: address,
	}
}

// newBucketRecord retorna um BucketRecord vazio
func newBucketRecord() BucketRecord {
	return BucketRecord{}
}

// ======================================= Crud ======================================== //

// HashCreate adiciona um novo Pokemon à estrutura de hash dinâmica e salva as alterações no arquivo.
// Recebe um Pokemon e a posição do registro no arquivo binário.
func HashCreate(pokemon models.Pokemon, address int64) {
	// Cria bucket, importa o diretorio, adiciona aos buckets e salva arquivo
	pokeRecord := pokemonToBucketRecord(pokemon, address)
	hash, _ := loadDinamicHash(DIRECTORY_FILE)
	hash.addRecord(pokeRecord)
	hash.Close()
}

// HashRead busca um Pokemon no arquivo binário usando a estrutura de hash dinâmica.
// Retorna o Pokemon encontrado, a posição do registro no arquivo e um erro, se houver.
func HashRead(targetID int64) (models.Pokemon, int64, error) {
	// Carrega o diretorio e o bucket para a memoria primaria
	hash, _ := loadDinamicHash(DIRECTORY_FILE)
	pos := targetID % int64(hash.getBucketCount())
	bucket := hash.readBucket(pos)

	// Recupera o endereço do ID
	targetPos := int64(-1)
	for i := int64(0); i < bucket.CurrentSize; i++ {
		if bucket.Records[i].ID == targetID {
			targetPos = bucket.Records[i].Address
			i = bucket.CurrentSize
		}
	}

	// Realiza a leitura no arquivo a partir do endereço
	c, err := inicializarControleLeitura(BIN_FILE)
	targetPokemon := c.ReadTarget(targetPos)

	// Trata possiveis erros
	if targetPos < 0 {
		err = fmt.Errorf("pokemon nao encontrado")
	} else if targetPokemon.Numero < 0 {
		err = fmt.Errorf("arquivo corrompido")
	}

	return targetPokemon, targetPos, err
}

// HashDelete remove um Pokémon com base no ID fornecido da estrutura de hash dinâmica
// e atualiza o arquivo Hash.
//
// Retorna um erro se o Pokémon não for encontrado.
func HashDelete(targetID int64) error {
	// Recuperar o bucket
	hash, _ := loadDinamicHash(DIRECTORY_FILE)
	pos := targetID % int64(hash.getBucketCount())
	bucket := hash.readBucket(pos)

	// Procurar o registro no bucket
	targetPos := int64(-1)
	for i := int64(0); i < bucket.CurrentSize; i++ {
		if bucket.Records[i].ID == targetID {
			targetPos = i
		}
	}
	// Se nao achar erro!
	if targetPos < 0 {
		return fmt.Errorf("valor nao encontrado")
	}

	// Remover o registro do bucket
	bucket.CurrentSize--
	for i := targetPos; i < int64(len(bucket.Records))-1; i++ {
		bucket.Records[i] = bucket.Records[i+1]
	}
	bucket.Records[len(bucket.Records)-1] = newBucketRecord()

	// Escreve novo bucket em arquivo
	hash.insertIntoBucket(pos, bucket.ActualPower, bucket.CurrentSize, bucket.Records)

	// Se estiver vazio atualiza o garbage collector
	if bucket.CurrentSize == 0 {
		bucket = Bucket{}
		hash.directory.garbage = append(hash.directory.garbage, hash.directory.bucketPointer[pos])
		hash.directory.bucketPointer[pos] = hash.directory.bucketPointer[pos>>1]
		mergedBucket := hash.readBucket(pos)
		mergedBucket.ActualPower--
		hash.insertIntoBucket(pos, mergedBucket.ActualPower, mergedBucket.CurrentSize, mergedBucket.Records)
	}

	// Salva o arquivo do diretorio
	hash.Close()

	return nil
}

// HashUpdate atualiza a localização de um Pokémon na estrutura de hash dinâmica,
// fornecendo o novo endereço. Retorna um erro se o Pokémon não for encontrado.
func HashUpdate(pokemon models.Pokemon, newAddress int64) error {
	// Carrega o diretorio e o bucket para a memoria primaria
	hash, _ := loadDinamicHash(DIRECTORY_FILE)
	pos := int64(pokemon.Numero) % int64(hash.getBucketCount())
	bucket := hash.readBucket(pos)

	// Recupera o endereço do ID
	targetPos := int64(-1)
	for i := int64(0); i < bucket.CurrentSize; i++ {
		if bucket.Records[i].ID == int64(pokemon.Numero) {
			targetPos = i
			i = bucket.CurrentSize
		}
	}

	// Trata possiveis erros
	if targetPos < 0 {
		return fmt.Errorf("pokemon nao encontrado para atualizar")
	}

	// Atualiza para o novo endereço
	bucket.Records[targetPos].Address = newAddress
	hash.insertIntoBucket(pos, bucket.ActualPower, bucket.CurrentSize, bucket.Records)

	// Fecha a hash de maneira segura
	hash.Close()

	return nil
}
