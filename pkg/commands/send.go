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
	"log"

	"github.com/mavjs/pushnotifier"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// sendCmd represents the send command
var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Sends different types of content to registered devices.",
	Long: `This commands allow you to send text, URL or both, image to your registered devices via pushnotifier.de
To send a text notification, you can invoke the command as: send "my text notification"
`,
	Run: func(cmd *cobra.Command, args []string) {

		if len(args) > 1 {
			cobra.CheckErr("Too many arguments. Provide only 1 argument to send as text content.")
		}

		textContent := cmd.Flags().Arg(0)

		notifySend, err := cmd.Flags().GetBool("notify")
		if err != nil {
			cobra.CheckErr(err)
		}

		devices, err := cmd.Flags().GetStringSlice("devices")
		if err != nil {
			cobra.CheckErr(err)
		}

		urlContent, err := cmd.Flags().GetString("url")
		if err != nil {
			cobra.CheckErr(err)
		}

		imagePath, err := cmd.Flags().GetString("image")
		if err != nil {
			cobra.CheckErr(err)
		}

		silentSend, err := cmd.Flags().GetBool("silent")
		if err != nil {
			cobra.CheckErr(err)
		}

		packageName := viper.GetString("PACKAGE_NAME")
		apiToken := viper.GetString("API_TOKEN")
		appToken := viper.GetString("APP_TOKEN")

		if packageName == "" || apiToken == "" {
			cobra.CheckErr(errors.New("no package name or api token can be found. please use `register` command to register"))
		}
		pn := pushnotifier.NewClient(nil, packageName, apiToken, appToken)

		if notifySend {
			if textContent == "" && urlContent == "" {
				cobra.CheckErr("notify send option was selected however text and or url content not provided")
			}

			log.Println("Sending notification with both text and url")
			if err := pn.SendNotification(textContent, urlContent, devices, silentSend); err != nil {
				cobra.CheckErr(err)
			}
		}

		if textContent != "" && urlContent == "" {
			log.Println("Sending text notification")
			if err := pn.SendText(textContent, devices, silentSend); err != nil {
				cobra.CheckErr(err)
			}
		}

		if textContent == "" && urlContent != "" {
			log.Println("Sending URL notification")
			if err := pn.SendURL(urlContent, devices, silentSend); err != nil {
				cobra.CheckErr(err)
			}
		}

		if imagePath != "" {
			log.Println("Sending image notification")
			if err := pn.SendImage(imagePath, devices, silentSend); err != nil {
				cobra.CheckErr(err)
			}
		}
	},
}

func init() {
	rootCmd.AddCommand(sendCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// sendCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// sendCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	sendCmd.Flags().BoolP("notify", "n", false, "Option to send a notification that contains both text and url. User is taken to URL when tapping notification")

	sendCmd.Flags().StringP("url", "u", "", "The URL to include during notification send")
	sendCmd.Flags().StringP("image", "i", "", "The path to an iamge to send as notification")

	sendCmd.Flags().StringSliceP("devices", "d", make([]string, 0), "List of device IDs to send notification")

	sendCmd.Flags().BoolP("silent", "s", false, "Option to send notification in silent mode")

}
