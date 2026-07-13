package git

import (
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"knotwork-core/internal/domain"
)

func normalize(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	return strings.ReplaceAll(text, "\r", "\n")
}

func splitBlocks(text string) [][]string {
	blockRegex := regexp.MustCompile(`\n[\n\s]*\n`)
	blocks := blockRegex.Split(text, -1)

	result := [][]string{}
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

func parseBlocks(blocks [][]string) ([]block, error) {
	blockData := []block{}
	for _, bl := range blocks {
		indent, err := getIndent(bl)
		if err != nil {
			return nil, err
		}

		if isCategoryBlock(bl) {
			hash := sha1.Sum([]byte(bl[1]))
			blockData = append(blockData, category{
				category: bl[1],
				id:       hex.EncodeToString(hash[:]),
			})
			continue
		}

		blockLines := []string{}
		for _, line := range bl {
			blockLines = append(blockLines, strings.TrimSpace(line))
		}

		finished, err := isFinished(blockLines)
		if err != nil {
			return nil, err
		}

		hash := sha1.Sum([]byte(blockLines[0]))
		blockData = append(blockData, task{
			level:    indent,
			title:    blockLines[0],
			updates:  blockLines[1:],
			id:       hex.EncodeToString(hash[:]),
			finished: finished,
		})
	}

	return blockData, nil
}

func validateHeirarchy(blockData []block) error {
	currIndents := []int{-1}
	currCategory := ""
	firstBlock := true

	for _, bl := range blockData {
		if categoryBlock, ok := bl.(category); ok {
			firstBlock = true
			currIndents = []int{-1}
			currCategory = categoryBlock.category
			continue
		}

		taskBlock, _ := bl.(task)

		level := taskBlock.level

		if firstBlock && level > 0 {
			return fmt.Errorf("invalid first task for %s", currCategory)
		}

		for len(currIndents) > 0 && currIndents[len(currIndents)-1] >= level {
			currIndents = currIndents[:len(currIndents)-1]
		}

		if level-1 != currIndents[len(currIndents)-1] {
			return fmt.Errorf(`invalid parent task for "%s"`, taskBlock.title)
		}

		currIndents = append(currIndents, level)
		firstBlock = false
	}

	return nil
}

func validateBlockData(blockData []block) error {
	if len(blockData) == 0 {
		return errors.New("empty todo")
	}

	err := validateHeirarchy(blockData)
	if err != nil {
		return err
	}

	return nil
}

func buildTaskMap(blockData []block) taskMap {
	result := make(taskMap)
	currCategory := ""
	categorySet := false

	dummyTask := task{
		id:       "",
		level:    -1,
		title:    "",
		updates:  []string{},
		finished: false,
	}
	currParents := []task{dummyTask}
	for _, bl := range blockData {
		if categoryBlock, ok := bl.(category); ok {
			currCategory = categoryBlock.category
			categorySet = true
			currParents = []task{dummyTask}
			continue
		}

		taskBlock, _ := bl.(task)

		currentTask := domain.Task{
			Title:    taskBlock.title,
			Updates:  taskBlock.updates,
			Finished: taskBlock.finished,
		}

		if categorySet {
			currentTask.Category = currCategory
		}

		for len(currParents) > 0 &&
			currParents[len(currParents)-1].level >= taskBlock.level {
			currParents = currParents[:len(currParents)-1]
		}

		h := sha1.New()
		for _, parent := range currParents[1:] {
			h.Write([]byte(parent.id))
		}
		h.Write([]byte(taskBlock.id))
		taskId := hex.EncodeToString(h.Sum(nil))
		currentTask.Id = domain.TaskId(taskId)

		parentTitles := []string{}
		finished := taskBlock.finished
		for _, parent := range currParents[1:] {
			parentTitles = append(parentTitles, parent.title)

			if parent.finished {
				finished = true
			}
		}
		currentTask.ParentTasks = parentTitles
		currentTask.Finished = finished

		result[domain.TaskId(taskId)] = currentTask
		currParents = append(currParents, taskBlock)
	}

	return result
}

func ParseTodo(text string) (taskMap, error) {
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
