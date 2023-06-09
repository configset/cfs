package operators

import (
	"cfs/exe"
	"cfs/system"
	"fmt"
	"github.com/rs/zerolog/log"
	"os"
)

type Command struct {
	Cmd        string `yaml:"cmd"`
	WorkingDir string `yaml:"dir"`
	OsLimits   string `yaml:"osLimits"`
	SetEnv     string `yaml:"setEnv"`
}

func (c *Command) Setup() {
	c.WorkingDir = RenderEnvString(c.WorkingDir)
}

// Execute runs the command.
func (c *Command) Execute() error {
	c.Setup()
	/* We do not replace command envars like the other functions, this is intended to be a raw command */
	if system.Get().CanExecOnOs(c.OsLimits) {
		if len(c.Cmd) < 1 {
			return fmt.Errorf("no command to execute")
		}
		logStr := c.Cmd
		if len(logStr) > 25 {
			logStr = logStr[0:25] + "..."
		}
		log.Info().Msgf("executing command: %s", logStr)
		fileName := exe.EchoToFile(c.Cmd, os.TempDir())
		// change directory to the working directory if specified
		err := os.Chmod(fileName, 0775)
		if err != nil {
			log.Error().Err(err).Msg("temp file must exist to continue")
			return err
		}
		log.Debug().Str("command", c.Cmd).Msgf("executing local file: %s", fileName)
		pc := exe.Run(fileName, c.WorkingDir)
		if pc.Failed() {
			log.Error().Err(pc.GetErr()).Msg(pc.Get())
			return pc.GetErr()
		} else {
			log.Debug().Str("cmd", c.Cmd).Msgf("completed executing: %s", fileName)
			log.Debug().Msgf("Output: %s", pc.Get())
			if len(c.SetEnv) > 0 {
				log.Debug().Str("cmd", c.Cmd).Msgf("setting env var: %s=%s", c.SetEnv, pc.Get())
				os.Setenv(c.SetEnv, pc.Get())
			}
			os.Remove(fileName)
		}
	} else {
		log.Info().Str("cmd", c.Cmd).Msgf("skipped due to os limit: %s", c.OsLimits)
	}
	return nil
}
