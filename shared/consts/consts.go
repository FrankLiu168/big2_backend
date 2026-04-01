package consts

type CardType int

const (
	CARD_TYPE_UNKNOWN  CardType = 0 //  未知
	CARD_TYPE_SINGLE   CardType = 1 //  單張
	CARD_TYPE_ONE_PAIR CardType = 2 //  對子

	CARD_TYPE_STRAIGHT       CardType = 3 //  順子
	CARD_TYPE_FULL_HOUSE     CardType = 4 //  葫蘆
	CARD_TYPE_FOUR_OF_A_KIND CardType = 5 //  鐵支
	CARD_TYPE_STRAIGHT_FLUSH CardType = 6 //  同花順
)

const (
	START_CARD = 103
)

// 牌型名稱對照表
var CARD_TYPE_MAP = map[CardType]string{
	CARD_TYPE_UNKNOWN:        "未知",
	CARD_TYPE_SINGLE:         "單張",
	CARD_TYPE_ONE_PAIR:       "對子",
	CARD_TYPE_STRAIGHT:       "順子",
	CARD_TYPE_FULL_HOUSE:     "葫蘆",
	CARD_TYPE_FOUR_OF_A_KIND: "鐵支",
	CARD_TYPE_STRAIGHT_FLUSH: "同花順",
}

// 花色名稱
var SUIT_NAME_MAP = map[int]string{
	1: "梅花",
	2: "方塊",
	3: "紅心",
	4: "黑桃",
}

// 點數名稱
var RANK_NAME_MAP = map[int]string{
	1:  "A",
	2:  "2",
	3:  "3",
	4:  "4",
	5:  "5",
	6:  "6",
	7:  "7",
	8:  "8",
	9:  "9",
	10: "10",
	11: "J",
	12: "Q",
	13: "K",
}

func mapingValueGetKey[T1 comparable, T2 comparable](m map[T1]T2, value T2) (T1, bool) {
	for k, v := range m {
		if v == value {
			return k, true
		}
	}
	var zero T1 // 零值
	return zero, false
}

func GetCardNumber(cardName string) int {
	runes := []rune(cardName)       
	suitName := string(runes[0:2])   
	rankName := string(runes[2:])   
	suit, _ := mapingValueGetKey(SUIT_NAME_MAP, suitName)
	rank, _ := mapingValueGetKey(RANK_NAME_MAP, rankName)
	return suit*100 + rank

}

func GetCardName(cardNo int) string {
	suit := cardNo / 100
	rank := cardNo % 100
	suitName := SUIT_NAME_MAP[suit]
	rankName := RANK_NAME_MAP[rank]
	return suitName + rankName
}

func GetCardNameList(cardNoList []int) []string {
	cardNameList := make([]string, len(cardNoList))
	for i, cardNo := range cardNoList {
		cardNameList[i] = GetCardName(cardNo)
	}
	return cardNameList
}

func GetCardNumberList(cardNameList []string) []int {
	cardNumberList := make([]int, len(cardNameList))
	for i, cardName := range cardNameList {
		cardNumberList[i] = GetCardNumber(cardName)
	}
	return cardNumberList
}

func GetCardTypeNo(cardType string) CardType {
	for k, v := range CARD_TYPE_MAP {
		if v == cardType {
			return k
		}
	}
	return CARD_TYPE_UNKNOWN
}

func GetCardTypeName(cardType CardType) string {
	return CARD_TYPE_MAP[cardType]
}
