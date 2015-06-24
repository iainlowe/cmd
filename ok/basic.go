package main

import (
	"flag"
	"fmt"
	"log"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

var (
	fst *string = flag.String("fst", "60:80", "filesystem threshold (warn above fst usage)")
)

func checkfs() error {
	findroots := exec.Command("cat", "/proc/mounts")
	b, _ := findroots.Output()
	seen := map[string]bool{}

	threshparts := strings.Split(*fst, ":")
	wthresh, _ := strconv.ParseInt(threshparts[0], 10, 32)
	cthresh, _ := strconv.ParseInt(threshparts[1], 10, 32)

	for _, root := range strings.Split(string(b), "\n") {
		rootdev := strings.Split(root, " ")[0]

		if _, ok := seen[rootdev]; !strings.HasPrefix(root, "/dev/") || ok {
			continue
		}

		root = strings.Split(root, " ")[1]
		root = strings.TrimSpace(root)

		stats := &syscall.Statfs_t{}

		if err := syscall.Statfs(root, stats); err != nil {
			// if strings.Contains(err.Error(), "permission denied") {
			// 	continue
			// }
			log.Fatal(err)
		}

		total := stats.Blocks * uint64(stats.Frsize)
		free := stats.Bavail * uint64(stats.Frsize)
		pct := 100.0 / float64(total) * (float64(total) - float64(free))

		seen[rootdev] = true

		if pct > float64(cthresh) {
			warn("disk", root, "on", "gecko", "is critical:", fmt.Sprintf("%2.2f", pct)+"%")
			continue
		} else if pct > float64(wthresh) {
			warn("disk", root, "on", "gecko", "is warning:", fmt.Sprintf("%2.2f", pct)+"%")
			continue
		}

		if *verbose {
			if *human {
				fmt.Printf("%s %s %s %2.2f", root, bytesToHuman(total), bytesToHuman(free), pct)
				fmt.Println("%")
			} else {
				fmt.Printf("%s %d %d %1.4f\n", root, total, free, pct/100)
			}
		}

	}

	return nil
}
