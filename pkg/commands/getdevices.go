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

	"github.com/mavjs/pushnotifier"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getdevicesCmd represents the getdevices command
var getdevicesCmd = &cobra.Command{
	Use:   "getdevices",
	Short: "Get connected devices",
	Long:  `Get connected devices to your account and show them for convenience and or later used for sending notifications.`,
	Run: func(cmd *cobra.Command, args []string) {
		packageName := viper.GetString("PACKAGE_NAME")
		apiToken := viper.GetString("API_TOKEN")
		appToken := viper.GetString("APP_TOKEN")

		if packageName == "" || apiToken == "" {
			cobra.CheckErr(errors.New("no package name or api token can be found. please use `register` command to register"))
		}

		pn := pushnotifier.NewClient(nil, packageName, apiToken, appToken)

		pn.GetDevices()

		for _, device := range pn.Devices {
			fmt.Println(device)
		}
	},
}

func init() {
	rootCmd.AddCommand(getdevicesCmd)
}
