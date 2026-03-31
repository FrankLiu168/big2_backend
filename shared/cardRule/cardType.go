package cardrule

import (
	"big2backend/shared/consts"
	"sort"
)

func GetCardType(cards []int) consts.CardType {
	switch true {
	case IsStraightFlush(cards):
		return consts.CARD_TYPE_STRAIGHT_FLUSH
	case IsFullHouse(cards):
		return consts.CARD_TYPE_FULL_HOUSE
	case IsStraight(cards):
		return consts.CARD_TYPE_STRAIGHT
	case IsPair(cards):
		return consts.CARD_TYPE_ONE_PAIR
	case IsSingle(cards):
		return consts.CARD_TYPE_SINGLE
	}
	return consts.CARD_TYPE_UNKNOWN
}

func IsSingle(cards []int) bool {
	if len(cards) != 1 {
		return false
	}
	return true
}

func IsPair(cards []int) bool {
	if len(cards) != 2 {
		return false
	}
	if cards[0]%100 != cards[1]%100 {
		return false
	}
	return true
}

func IsStraight(cards []int) bool {
	if len(cards) != 5 {
		return false
	}
	nums := []int{}
	for _, card := range cards {
		nums = append(nums, card%100)
	}
	flag := true
	sort.Ints(nums)
	s1 := []int{1, 10, 11, 12, 13}
	for i := range s1 {
		if nums[i] != s1[i] {
			flag = false
			break
		}
	}
	if flag {
		return true
	}

	for i := range len(nums) - 1 {
		if nums[i]+1 != nums[i+1] {
			return false
		}
	}
	return true
}

func IsFullHouse(cards []int) bool {
	if len(cards) != 5 {
		return false
	}
	var nums = []int{}
	for _, card := range cards {
		nums = append(nums, card%100)
	}
	sort.Ints(nums)
	var numMap = map[int]int{}
	for i := range len(nums) - 1 {
		num := nums[i]
		_, isExit := numMap[num]
		if isExit {
			numMap[num] += 1
		} else {
			numMap[num] = 1
		}
	}
	if len(numMap) != 2 {
		return false
	}
	for _, count := range numMap {
		if count != 3 && count != 2 {
			return false
		}
	}
	return true
}

func IsFourOfAKind(cards []int) bool {
	if len(cards) != 5 {
		return false
	}
	nums := []int{}
	for i := range cards {
		nums = append(nums, cards[i]%4)
	}
	numMap := map[int]int{}
	for i := range len(nums) - 1 {
		num := nums[i]
		_, isExit := numMap[num]
		if isExit {
			numMap[num] += 1
		} else {
			numMap[num] = 1
		}
	}
	if len(numMap) != 2 {
		return false
	}
	for _, count := range numMap {
		if count != 4 && count != 1 {
			return false
		}
	}
	return true
}

func isFlush(cards []int) bool {
	nums := []int{}
	for i := range cards {
		nums = append(nums, cards[i]/100)
	}
	for i := range len(nums) - 1 {
		if nums[i] != nums[i+1] {
			return false
		}
	}
	return true
}

func IsStraightFlush(cards []int) bool {
	return IsStraight(cards) && isFlush(cards)
}
