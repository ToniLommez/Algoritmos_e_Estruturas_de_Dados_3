package sorts

import "github.com/Bernardo46-2/AEDS-III/data/binManager"

const FILE string = binManager.CSV_PATH
const BIN_FILE string = binManager.BIN_FILE

type SortFunc func()

var SortingFunctions = []SortFunc{
	IntercalacaoBalanceadaComum,
	IntercalacaoBalanceadaVariavel,
	IntercalacaoPorSubstituicao,
}