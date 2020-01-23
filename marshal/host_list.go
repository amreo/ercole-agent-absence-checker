// Copyright (c) 2019 Sorint.lab S.p.A.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package marshal

import (
	"bufio"
	"strings"
)

// Oratab marshals a list of dbs (one per line)
func HostList(cmdOutput []byte) []string {
	list := make([]string, 0)
	scanner := bufio.NewScanner(strings.NewReader(string(cmdOutput)))
	for scanner.Scan() {
		line := scanner.Text()
		list = append(list, strings.TrimSpace(line))
	}

	return list
}
