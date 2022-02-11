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

import model "apigateway/core/model/laundry"

// Services exposes APIs for the driver adapters
type Services interface {
	GetVersion() string
	StoreRecord(name string) error
}

type servicesImpl struct {
	app *Application
}

func (s *servicesImpl) GetVersion() string {
	return s.app.getVersion()
}

func (s *servicesImpl) StoreRecord(name string) error {
	return s.app.storeRecord(name)
}

// Storage is used by core to storage data - DB storage adapter, file storage adapter etc
type Storage interface {
	StoreRecord(name string) error
}

type Laundry interface {
	ListRooms() (*model.Organization, error)
	GetLaundryRoom(roomid int) (*model.RoomDetail, error)
}
