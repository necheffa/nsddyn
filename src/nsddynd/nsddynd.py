"""
    Copyright (C) 2019 Alexander Necheff

    This program is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, version 3 of the License.

    This program is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.

    You should have received a copy of the GNU General Public License
    along with this program.  If not, see <https://www.gnu.org/licenses/>.
"""

import localzone

# TODO: this needs to be taken from a configuration instead of hard coded.
__origin="dyn.necheff.net"

def update_record(zone, host, ipaddr):
    """
    update_record() searches the specified zone for a given A record and attempts to replace
    it with the given IP address. If the record does not exist then updated_record() just
    appends the zone with a new A record using the given IP address.
    """

    dynzone = localzone.context.load(zone,origin=__origin)
    records = dynzone.find_record(rdtype="A", name=host)
    if len(records) > 0:
        # the hostname exists in the zone file so remove it before adding a new one
        for r in records:
            dynzone.remove_record(r)

    dynzone.add_record(host, "A", ipaddr)
    dynzone.save()

if __name__ == "__main__":
    update_record("test.zone", "myhost", "192.0.2.2")

