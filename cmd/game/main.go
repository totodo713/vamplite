package main

import (
	    "log"
	        "muscle-dreamer/internal/core"
	)

	func main() {
		    game := core.NewGame()
		        if err := game.Run(); err != nil {
				        log.Fatal(err)
					    }
				    }
