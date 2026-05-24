package main

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

func normalize(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}

func splitBlocks(text string) [][]string {
	blockRegex := regexp.MustCompile(`\n[\n\s]*\n`)
	blocks := blockRegex.Split(text, -1)

	var result [][]string
	for _, block := range blocks {
		lines := strings.Split(block, "\n")
		result = append(result, lines)
	}
	return result
}

func isCategoryBlock(block []string) bool {
	if len(block) != 3 {
		return false
	}

	border := strings.Repeat("*", 32)
	return block[0] == border && block[2] == border
}

func getIndent(block []string) (int, error) {
	var currIndent string
	hasIndent := false
	indentPattern := regexp.MustCompile(`^((?: {4})*)\S.*$`)
	for _, line := range block {
		matches := indentPattern.FindStringSubmatch(line)
		if matches == nil {
			return -1, fmt.Errorf("invalid indentation for %q", line)
		}

		indent := matches[1]
		if !hasIndent {
			currIndent = indent
			hasIndent = true
		} else if currIndent != indent {
			return -1, fmt.Errorf("inconsistent indentation for %q", indent)
		}
	}

	return len(currIndent) / 4, nil
}

func isFinished(blockLines []string) (bool, error) {
	matched, err := regexp.MatchString(`^\[DONE\].*$`, blockLines[len(blockLines)-1])
	if err != nil {
		return false, err
	}

	return len(blockLines) >= 2 && matched, nil
}

func parseBlocks(blocks [][]string) ([]map[string]any, error) {
	var blockData []map[string]any
	for _, block := range blocks {
		indent, err := getIndent(block)
		if err != nil {
			return nil, err
		}

		if isCategoryBlock(block) {
			hash := sha1.Sum([]byte(block[1]))
			blockData = append(blockData, map[string]any{
				"category": block[1],
				"id":       hex.EncodeToString(hash[:]),
			})
			continue
		}

		var blockLines []string
		for _, line := range block {
			blockLines = append(blockLines, strings.TrimSpace(line))
		}


		finished, err := isFinished(blockLines)
		if err != nil {
			return nil, err
		}

		hash := sha1.Sum([]byte(blockLines[0]))
		blockData = append(blockData, map[string]any{
			"level":    indent,
			"title":    blockLines[0],
			"updates":  blockLines[1:],
			"id":       hex.EncodeToString(hash[:]),
			"finished": finished,
		})
	}

	return blockData, nil
}

func validateHeirarchy(blockData []map[string]any) error {
	currIndents := []int{-1}
	currCategory := ""
	firstBlock := true

	for _, block := range blockData {
		if category, ok := block["category"]; ok {
			firstBlock = true
			currIndents = []int{-1}
			currCategory = category.(string)
			continue
		}

		level := block["level"].(int)

		if firstBlock && level > 0 {
			return fmt.Errorf("invalid first task for %s", currCategory)
		}

		for len(currIndents) > 0 && currIndents[len(currIndents)-1] >= level {
			currIndents = currIndents[:len(currIndents)-1]
		}

		if level-1 != currIndents[len(currIndents)-1] {
			return fmt.Errorf(`invalid parent task for "%s"`, block["title"].(string))
		}

		currIndents = append(currIndents, level)
		firstBlock = false
	}

	return nil
}

func validateBlockData(blockData []map[string]any) error {
	if len(blockData) == 0 {
		return errors.New("empty todo")
	}

	err := validateHeirarchy(blockData)
	if err != nil {
		return err
	}

	return nil
}

func buildTaskMap(blockData []map[string]any) map[string]any {
	result := make(map[string]any)
	var currCategory string
	categorySet := false

	dummyTask := map[string]any{
		"level": -1,
		"id":    "",
	}
	currParents := []map[string]any{dummyTask}
	for _, block := range blockData {
		if category, ok := block["category"].(string); ok {
			currCategory = category
			categorySet = true
			currParents = []map[string]any{dummyTask}
			continue
		}

		currentTask := map[string]any{
			"title":    block["title"],
			"updates":  block["updates"],
			"finished": block["finished"],
		}

		if categorySet {
			currentTask["category"] = currCategory
		}

		for len(currParents) > 0 &&
			currParents[len(currParents)-1]["level"].(int) >= block["level"].(int) {
			currParents = currParents[:len(currParents)-1]
		}

		h := sha1.New()
		for _, parent := range currParents[1:] {
			h.Write([]byte(parent["id"].(string)))
		}
		h.Write([]byte(block["id"].(string)))
		taskId := hex.EncodeToString(h.Sum(nil))
		currentTask["id"] = taskId

		parentTitles := []string{}
		finished := block["finished"].(bool)
		for _, parent := range currParents[1:] {
			parentTitles = append(parentTitles, parent["title"].(string))

			if pf, ok := parent["finished"].(bool); ok && pf {
				finished = true
			}
		}
		currentTask["parentTasks"] = parentTitles
		currentTask["finished"] = finished

		result[taskId] = currentTask
		currParents = append(currParents, block)
	}

	return result
}

func ParseTodo(text string) (map[string]any, error) {
	text = normalize(text)
	blocks := splitBlocks(text)

	blockData, err := parseBlocks(blocks)
	if err != nil {
		return nil, err
	}

	err = validateBlockData(blockData)
	if err != nil {
		return nil, err
	}

	return buildTaskMap(blockData), nil
}
