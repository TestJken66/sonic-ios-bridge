/*
 *  Copyright (C) [SonicCloudOrg] Sonic Project
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *         http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 *
 */
package cmd

import (
	"fmt"
	"github.com/SonicCloudOrg/sonic-ios-bridge/src/util"
	giDevice "github.com/electricbubble/gidevice"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
)

var syslogCmd = &cobra.Command{
	Use:   "syslog",
	Short: "Get syslog from your device.",
	Long:  "Get syslog from your device.",
	RunE: func(cmd *cobra.Command, args []string) error {
		usbMuxClient, err := giDevice.NewUsbmux()
		if err != nil {
			return util.NewErrorPrint(util.ErrConnect, "usbMux", err)
		}
		list, err1 := usbMuxClient.Devices()
		if err1 != nil {
			return util.NewErrorPrint(util.ErrSendCommand, "listDevices", err1)
		}
		if len(list) == 0 {
			fmt.Println("no device connected")
			os.Exit(0)
		} else {
			var device giDevice.Device
			if len(udid) != 0 {
				for i, d := range list {
					if d.Properties().SerialNumber == udid {
						device = list[i]
						break
					}
				}
			} else {
				device = list[0]
			}
			output, err := device.Syslog()
			if err != nil {
				fmt.Printf("Get syslog failed: %s", err)
			}
			defer device.SyslogStop()
			done := make(chan os.Signal, syscall.SIGTERM)
			signal.Notify(done, os.Interrupt, os.Kill)

			go func() {
				for line := range output {
					if len(filter) == 0 {
						fmt.Print(line)
						continue
					} else {
						if strings.Contains(line, filter) {
							fmt.Print(line)
						}
					}
				}
				done <- os.Interrupt
			}()

			<-done
		}
		return nil
	},
}

var filter string

func init() {
	rootCmd.AddCommand(syslogCmd)
	syslogCmd.Flags().StringVarP(&udid, "udid", "u", "", "device's serialNumber ( default first device )")
	syslogCmd.Flags().StringVarP(&filter, "filter", "f", "", "filter by some message.")
}
