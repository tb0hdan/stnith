package engine

import (
	"fmt"
	"log"
	"os"

	"github.com/tb0hdan/stnith/pkg/engine/destructors"
	"github.com/tb0hdan/stnith/pkg/engine/disablers"
	"github.com/tb0hdan/stnith/pkg/engine/failsafes"
	"github.com/tb0hdan/stnith/pkg/engine/savers"
)

type EngineInterface interface {
	Run()
}

type Engine struct {
	disablerList   []disablers.DisablerInterface
	destructorList []destructors.DestructorInterface
	failsafeList   []failsafes.FailsafeInterface
	saverList      []savers.SaverInterface
}

func New(disablerList []disablers.DisablerInterface, destructorList []destructors.DestructorInterface,
	failsafeList []failsafes.FailsafeInterface, saverList []savers.SaverInterface) EngineInterface {
	return &Engine{
		disablerList:   disablerList,
		destructorList: destructorList,
		failsafeList:   failsafeList,
		saverList:      saverList,
	}
}

func (e *Engine) Run() {
	fmt.Println("\nTimer expired!")

	// First, trigger failsafeList to hide the process if configured
	if len(e.failsafeList) > 0 {
		fmt.Println("Triggering failsafeList...")
		for _, failsafe := range e.failsafeList {
			if err := failsafe.Trigger(); err != nil {
				log.Printf("FailsafeInterface error: %v", err)
			}
		}
		fmt.Println("All failsafeList have been triggered.")
	}

	// Second, execute saverList if any are configured
	if len(e.saverList) > 0 {
		fmt.Println("Calling saverList...")
		for _, saver := range e.saverList {
			if err := saver.Save(); err != nil {
				log.Printf("SaverInterface error: %v", err)
			}
		}
		fmt.Println("All saverList have been called.")
	}

	// Third, disable security systems if any disablerList are configured
	if len(e.disablerList) > 0 {
		fmt.Println("Calling disablerList...")
		for _, disabler := range e.disablerList {
			if err := disabler.Disable(); err != nil {
				log.Printf("DisablerInterface error: %v", err)
			}
		}
		fmt.Println("All disablerList have been called.")
	}

	// Finally, execute destructorList
	if len(e.destructorList) == 0 {
		fmt.Println("No destructorList defined, exiting.")
		os.Exit(0)
	}
	fmt.Println("Calling destructorList...")
	for _, destructor := range e.destructorList {
		if err := destructor.Destroy(); err != nil {
			log.Printf("DestructorInterface error: %v", err)
		}
	}
	fmt.Println("All destructorList have been called. Good luck.")
	os.Exit(0)
}
