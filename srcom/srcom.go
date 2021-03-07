package srcom

import (
	"errors"
	"strings"

	"github.com/Witcher01/srapi"
)

func GetTopRunnersByGame(id string, category string, amount int, platform string) (string, error) {
	collection, err := getLeaderboard(id, category, amount, platform)
	if err != nil {
		return "", err
	}

	var players_sb strings.Builder
	for _, run := range collection.Runs {
		players, err := run.Run.Players()
		if err != nil {
			return "", err
		}

		for _, player := range players.Players() {
			players_sb.WriteString(player.Name())
			players_sb.WriteString(", ")
		}
	}
	ret := players_sb.String()
	ret = strings.TrimSuffix(ret, ", ")

	return ret, nil
}

func GetTopByGame(id string, category string, amount int, platform string) (string, error) {
	collection, err := getLeaderboard(id, category, amount, platform)
	if err != nil {
		return "", err
	}

	var pbs strings.Builder
	for _, run := range collection.Runs {
		time, _err := getFormattedPlayerTime(run)
		if err != nil {
			return "", _err
		}

		pbs.WriteString(time)
		pbs.WriteRune('\n')
	}

	return pbs.String(), nil
}

func GetWRByGame(id string, category string, platform string) (string, error) {
	collection, err := getLeaderboard(id, category, 1, platform)
	if err != nil {
		return "", err
	}

	wr := collection.Runs[0]

	time, _err := getFormattedPlayerTime(wr)
	if _err != nil {
		return "", err
	}

	return time, nil
}

func getFormattedTime(run srapi.RankedRun) string  {
	times := run.Run.Times
	var time string
	if times.IngameTime.String() != "0s" {
		time = run.Run.Times.IngameTime.Format() + " (IGT)"
	} else if times.RealtimeWithoutLoads.String() != "0s" {
		time = run.Run.Times.RealtimeWithoutLoads.Format() + " (RTA/NL)"
	} else {
		time = run.Run.Times.Realtime.Format() + " (RTA)"
	}

	return time
}

func getFormattedPlayerTime(run srapi.RankedRun) (string, error) {
	players, err := run.Run.Players()
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, player := range players.Players() {
		sb.WriteString(player.Name())
		sb.WriteString(", ");
	}
	sb_string := sb.String()
	sb_string = strings.TrimSuffix(sb_string, ", ")

	time := getFormattedTime(run)

	sb.Reset()
	sb.WriteString(sb_string)
	sb.WriteString(": ")
	sb.WriteString(time)

	return sb.String(), nil
}

func getDefaultGameLeaderboard(id string, amount int, platform string) (*srapi.Leaderboard, error) {
	game, err := srapi.GameByID(id, srapi.NoEmbeds)
	if err != nil {
		return nil, err
	}

	filter := &srapi.LeaderboardOptions{Top: amount, Platform: platform}

	collection, err := game.PrimaryLeaderboard(filter, srapi.NoEmbeds)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

func getGameLeaderboardByCategory(id string, category string, amount int, platform string) (*srapi.Leaderboard, error) {
	embeds := srapi.NoEmbeds
	if category != "" {
		embeds = category
	}

	game, err := srapi.GameByID(id, embeds)
	if err != nil {
		return nil, err
	}

	cats, err := game.Categories(nil, nil, srapi.NoEmbeds)
	if err != nil {
		return nil, err
	}

	var cat *srapi.Category = nil
	cats.Walk(func(c *srapi.Category) bool {
		if (strings.ToLower(c.Name) == strings.ToLower(category)) {
			cat = c
			return false
		}
		return true
	});

	if cat == nil {
		return nil, errors.New("No category with that name found.")
	}

	filter_options := &srapi.LeaderboardOptions{Top: amount, Platform: platform}

	collection, err := cat.PrimaryLeaderboard(filter_options, srapi.NoEmbeds)
	if err != nil {
		return nil, err
	}

	return collection, nil
}

func getLeaderboard(id string, category string, amount int, platform string) (*srapi.Leaderboard, error) {
	if category != "" {
		return getGameLeaderboardByCategory(id, category, amount, platform)
	}

	return getDefaultGameLeaderboard(id, amount, platform)
}
