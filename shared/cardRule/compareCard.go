package cardrule

import "sort"

func CompareSingle(cards1 []int, cards2 []int) bool {
	rank1 := cards1[0] % 100
	rank2 := cards2[0] % 100
	if rank2 == rank1 {
		suit1 := cards1[0] / 100
		suit2 := cards2[0] / 100
		return suit2 > suit1
	}
	if rank2 == 2 {
		return true
	}
	if rank2 == 1 && rank1 != 2 {
		return true
	}
	if rank2 > rank1 && rank1 != 1 && rank1 != 2 {
		return true
	}
	return false
}

func ComparePair(cards1 []int, cards2 []int) bool {
	sort.Ints(cards1)
	sort.Ints(cards2)
	rank1 := cards1[1] % 100
	rank2 := cards2[1] % 100
	if rank2 == rank1 {
		return cards2[0] > cards1[0]
	}
	return CompareSingle(cards1, cards2)
}

func CompareStraight(cards1 []int, cards2 []int) bool {
	sort.Ints(cards1)
	sort.Ints(cards2)
	rank1s := cards1[0] % 100
	rank1e := cards1[4] % 100
	rank2s := cards2[0] % 100
	rank2e := cards2[4] % 100
	if rank2s == rank1s {
		if rank2s == 2 {
			return cards2[0] > cards1[0]
		}
		if rank2s == 1 && rank2e == 5 {
			return cards2[4] > cards1[4]
		}
		if rank2s == 1 && rank2e == 13 {
			return cards2[0] > cards1[0]
		}
		return cards2[4] > cards1[4]
	}
	if rank2s == 2 {
		return true
	}
	if rank2s == 1 && rank2e == 13 && rank1s != 2 {
		return true
	}
	if rank1s == 2 {
		return false
	}
	if rank1s == 1 && rank1e == 13 && rank2s != 2 {
		return false
	}
	return rank2e > rank1e
}

func CompareFullHouse(cards1 []int, cards2 []int) bool {
	map1 := map[int]int{}
	m1Max := 0
	map2 := map[int]int{}
	m2Max := 0
	for i := range cards1 {
		val, isExist := map1[cards1[i]%100]
		if isExist {
			map1[cards1[i]%100] = val + 1
			if val+1 >= 3 {
				m1Max = cards1[i] % 100
			}
		} else {
			map1[cards1[i]%100] = 1
		}
	}
	for i := range cards2 {
		val, isExist := map2[cards2[i]%100]
		if isExist {
			map2[cards2[i]%100] = val + 1
			if val+1 >= 3 {
				m1Max = cards2[i] % 100
			}
		} else {
			map2[cards2[i]%100] = 1
		}
	}
	return m2Max > m1Max
}

func CompareFourOfAKind(cards1 []int, cards2 []int) bool {
	return CompareFullHouse(cards1, cards2)
}

func CompareStraightFlush(cards1 []int, cards2 []int) bool {
	return CompareStraight(cards1, cards2)
}
