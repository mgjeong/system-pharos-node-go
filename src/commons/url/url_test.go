/*******************************************************************************
 * Copyright 2017 Samsung Electronics All Rights Reserved.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 *******************************************************************************/
package url

import "fmt"

func ExampleBase() {
	fmt.Println(Base())
	// Output: /api/v1
}
func ExamplePharosAnchor() {
	fmt.Println(PharosAnchor())
	// Output: /pharos-anchor
}
func ExampleManagement() {
	fmt.Println(Management())
	// Output: /management
}
func ExampleMonitoring() {
	fmt.Println(Monitoring())
	// Output: /monitoring
}
func ExampleDeploy() {
	fmt.Println(Deploy())
	// Output: /deploy
}
func ExampleApps() {
	fmt.Println(Apps())
	// Output: /apps
}
func ExampleUpdate() {
	fmt.Println(Update())
	// Output: /update
}
func ExampleEvents() {
	fmt.Println(Events())
	// Output: /events
}
func ExampleStart() {
	fmt.Println(Start())
	// Output: /start
}
func ExampleStop() {
	fmt.Println(Stop())
	// Output: /stop
}
func ExampleRegister() {
	fmt.Println(Register())
	// Output: /register
}
func ExampleUnregister() {
	fmt.Println(Unregister())
	// Output: /unregister
}
func ExampleNodes() {
	fmt.Println(Nodes())
	// Output: /nodes
}
func ExamplePing() {
	fmt.Println(Ping())
	// Output: /ping
}
func ExampleDisk() {
	fmt.Println(Disk())
	// Output: /disk
}
func ExampleResource() {
	fmt.Println(Resource())
	// Output: /resource
}
func ExamplePerformance() {
	fmt.Println(Performance())
	// Output: /performance
}
func ExampleConfiguration() {
	fmt.Println(Configuration())
	// Output: /configuration
}
func ExampleDevice() {
	fmt.Println(Device())
	// Output: /device
}
func ExampleReboot() {
	fmt.Println(Reboot())
	// Output: /reboot
}
func ExampleRestore() {
	fmt.Println(Restore())
	// Output: /restore
}
func ExampleControl() {
	fmt.Println(Control())
	// Output: /control
}
func ExampleNotification() {
	fmt.Println(Notification())
	// Output: /notification
}
func ExampleWatch() {
	fmt.Println(Watch())
	// Output: /watch
}
