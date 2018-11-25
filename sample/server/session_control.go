/*
 * ORBIT - Interlink Remote Applications
 * Copyright (C) 2018  Roland Singer <roland.singer[at]desertbit.com>
 * Copyright (C) 2018  Sebastian Borchers <sebastian[at]desertbit.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package main

import (
	"github.com/desertbit/orbit/control"
	"github.com/desertbit/orbit/sample/api"
)

func (s *Session) ClientInfo() (info api.ClientInfoRet, err error) {
	ctx, err := s.ctrl.Call(api.ControlClientInfo, nil)
	if err != nil {
		return
	}

	err = ctx.Decode(&info)
	if err != nil {
		return
	}

	return
}

func (s *Session) serverInfo(ctx *control.Context) (v interface{}, err error) {
	v = api.ServerInfoRet{
		RemoteAddr: s.LocalAddr().String(),
		Uptime: s.server.uptime,
		ClientsCount: len(s.server.Sessions()),
	}
	return
}
