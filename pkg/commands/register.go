/*
Copyright © 2022 Maverick Kaung <mavjs01@gmail.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/
package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/mavjs/pushnotifier/pkg/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
)

// registerCmd represents the register command
var registerCmd = &cobra.Command{
	Use:   "register",
	Short: "Registers API authentication details",
	Long:  `Provide several API authentication related details in order for the script to work without having to enter them again.`,
	Run: func(cmd *cobra.Command, args []string) {
		if !term.IsTerminal(int(os.Stdout.Fd())) {
			cobra.CheckErr(errors.New("unknown terminal"))
		}

		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			cobra.CheckErr(err)
		}

		terminal := term.NewTerminal(os.Stdin, "")

		fmt.Print("Please register your API authentication details. For more info: https://api.pushnotifier.de/v2/doc/\n\r\n\r")
		fmt.Print("Enter your application package name: ")
		packageName, err := terminal.ReadLine()
		if err != nil {
			cobra.CheckErr(err)
		}
		viper.Set("PACKAGE_NAME", packageName)

		appToken, err := terminal.ReadPassword("APP Token: ")
		if err != nil {
			cobra.CheckErr(err)
		}
		viper.Set("APP_TOKEN", appToken)

		apiToken, err := terminal.ReadPassword("API Token: ")
		if err != nil {
			cobra.CheckErr(err)
		}
		viper.Set("API_TOKEN", apiToken)

		defer term.Restore(int(os.Stdin.Fd()), oldState)

		viper.SetConfigFile(config.GetConfigFilePath())

		fmt.Printf("writing authentication details to config: %v\n\r", viper.ConfigFileUsed())
		err = viper.WriteConfig()
		cobra.CheckErr(err)

	},
}

func init() {
	rootCmd.AddCommand(registerCmd)
}
