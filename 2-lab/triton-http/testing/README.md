# Testing

All testing is done using simple python functions. `tests.py` has the functions
that are the test cases. Each function is a unique test case, the doc string
describes what the function is testing. Each function starts with a `test_`
prefix. There are about 15 test cases in total.

The `request_manager.py` file has the RequestManager class which deals with all
the TCP and HTTP stuff, this is a utility class.

The `data.py` file has the data of the files we're testing in binary format.
The contents are used to compare with what the server sends.

## Useage
You need to have python3 installed to run the tests. You also need to have the
http-parser package installed for this. Try -
```
pip3 install http-parser
```

You can just run tests.py file using something like `pytest` or just open a
terminal and import the `tests.py` module. For example -

```
from tests import *
test_200_text()
```

If the function (test) runs without any `AssertionError` exceptions, then it
implies that the test has passed.
