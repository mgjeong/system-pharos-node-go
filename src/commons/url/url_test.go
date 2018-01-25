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
