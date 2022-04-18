package main

import (
	"log"
	"strings"

	"github.com/NaturalSelectionLabs/RSS3-PreGod/indexer/pkg/api/gitcoin"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/reptile/pkg/handler"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/database"
	"github.com/NaturalSelectionLabs/RSS3-PreGod/shared/pkg/logger"
)

func init() {
	if err := database.Setup(); err != nil {
		log.Fatalf("database.Setup err: %v", err)
	}
}

// pull information from url
/*
func main() {
	lastPos := handler.GetLastPostion()
	currentPos := lastPos + 1
	logger.Infof("lastPos:%d", lastPos)

	for {
		getProject := false
		getProjectStr := "false"
		netTag := true

		projectInfo, err := handler.GetResult(currentPos)
		if err != nil {
			if err.Error() == "get result false:StatusCode [403]" {
				netTag = false

				goto END
			}

			logger.Fatal(err)
		}

		if projectInfo != nil && err == nil {
			err = handler.SetResultInDB(projectInfo)
			if err != nil {
				logger.Fatal(err)
			}

			getProject = true
		}

		err = handler.SetLastPostion(currentPos)
		if err != nil {
			logger.Fatal(err)
		}

		if getProject == true {
			getProjectStr = "true"
		}

	END:
		log.Printf("get [%d] project info stage: %s", currentPos, getProjectStr)
		log.Printf("current position: %d", currentPos)
		log.Printf("------------------------------------------\n")

		if lastPos == 6000 {
			break
		}

		if netTag == true {
			lastPos = currentPos
			currentPos = lastPos + 1
		}

		time.Sleep(50 * time.Millisecond)
	}
}*/

// Change all adminaddress to lowercase
func main() {
	rangeMax := 20
	resultcount := int(handler.GetResultTotal())

	lastpos := handler.GetLastPostion(
		handler.GitcoinProjectIdentity2,
		handler.GitcoinProjectNetworkID2)

	if lastpos >= resultcount {
		logger.Infof("lastPos[%d] is the latest pos", lastpos)

		return
	}

	nextPos := lastpos + rangeMax
	if nextPos >= resultcount {
		nextPos = resultcount
	}

	projects, err := handler.GetResultFromDB(lastpos, nextPos)
	if err != nil {
		logger.Fatal(err)
	}

	processingProject := []gitcoin.ProjectInfo{}

	for _, project := range projects {
		project.AdminAddress = strings.ToLower(project.AdminAddress)
		processingProject = append(processingProject, project)
	}

	if len(processingProject) > 0 {
		err = handler.SetResultsInDB(processingProject)
		if err != nil {
			logger.Fatal(err)
		}
	}

	err = handler.SetLastPostion(
		nextPos,
		handler.GitcoinProjectIdentity2,
		handler.GitcoinProjectNetworkID2)
	if err != nil {
		logger.Fatal(err)
	}
}
