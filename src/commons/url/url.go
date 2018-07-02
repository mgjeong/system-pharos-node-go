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

// Package commons/url defines url used by Pharos Node.
package url

// Returning base url as string.
func Base() string { return "/api/v1" }

// Returning pharos anchor url as string.
func PharosAnchor() string { return "/pharos-anchor" }

// Returning management url as string.
func Management() string { return "/management" }

// Returning monitoring url as string.
func Monitoring() string { return "/monitoring" }

// Returning deploy url as string.
func Deploy() string { return "/deploy" }

// Returning Apps url as string.
func Apps() string { return "/apps" }

// Returning Update url as string.
func Update() string { return "/update" }

// Returning Events url as string.
func Events() string { return "/events" }

// Returning Start url as string.
func Start() string { return "/start" }

// Returning Stop url as string.
func Stop() string { return "/stop" }

// Returning Nodes url as string.
func Nodes() string { return "/nodes" }

// Returning Ping url as string.
func Ping() string { return "/ping" }

// Returning Register url as string.
func Register() string { return "/register" }

// Returning Unregister url as string.
func Unregister() string { return "/unregister" }

// Returning Resoucres url as string.
func Resource() string { return "/resource" }

// Returning Performance url as string.
func Performance() string { return "/performance" }

// Returning Configuration url as string.
func Configuration() string { return "/configuration" }

// Returning Device url as string.
func Device() string { return "/device" }

// Returning Reboot url as string.
func Reboot() string { return "/reboot" }

// Returning Restore url as string.
func Restore() string { return "/restore" }

// Returning Control url as string.
func Control() string { return "/control" }

//Returning Notification url as string.
func Notification() string { return "/notification" }

//Returning Watch url as string.
func Watch() string { return "/watch" }

//Returning Disk url as string.
func Disk() string { return "/disk" }
