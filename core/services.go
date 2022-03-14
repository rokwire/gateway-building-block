/*
 *   Copyright (c) 2020 Board of Trustees of the University of Illinois.
 *   All rights reserved.

 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at

 *   http://www.apache.org/licenses/LICENSE-2.0

 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package core

import laundry "apigateway/core/model/laundry"

func (app *Application) getVersion() string {
	return app.version
}

func (app *Application) storeRecord(name string) error {
	return app.storage.StoreRecord(name)
}

func (app *Application) listLaundryRooms() (laundry.Organization, error) {
	lr, _ := app.laundry.ListRooms()
	return *lr, nil
}

func (app *Application) listAppliances(id string) (laundry.RoomDetail, error) {
	ap, _ := app.laundry.GetLaundryRoom(id)
	return *ap, nil
}

func (app *Application) initServiceRequest(machineid string) (laundry.MachineRequestDetail, error) {
	sr, _ := app.laundry.InitServiceRequest(machineid)
	return *sr, nil
}

func (app *Application) submitServiceRequest(machineID string, problemCode string, comments string, firstname string, lastname string, phone string) (laundry.ServiceRequestResult, error) {
	srr, _ := app.laundry.SubmitServiceRequest(machineID, problemCode, comments, firstname, lastname, phone)
	return *srr, nil
}
