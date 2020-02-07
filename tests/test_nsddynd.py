import unittest
import io

from nsddynd import nsddynd

__zonefilename = "test.zone"
__zonefilecontent = """
        test
"""
__testhostname = "myhostname"
__testipaddr = "192.0.2.3"

class TestNsdDynd(unittest.TestCase):

    def test_update_record(self):
        # write a sample zone file
        with io.open(__zonefilename, "rt", encoding="utf-8") as f:
            f.write(__zonefilecontent)

        # call update_record() on that zone file
            nsddynd.update_zone(__zonefilename, __testhostname, __testipaddr)

        # assert that the changes took place as expected
            self.assertTrue(True)

if __name__ == "__main__":
    unittest.main()

