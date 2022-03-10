def foo():
  print("Call sonobuoy.startTest to name the current test.")
  sonobuoy.startTest("Test #1")
  x = 1
  print("Any failure will automatically fail the currently running test.")
  assert.equals(1,1,"X should be equal to %v but I got %v")
  print("Call sonobuoy.passTest to mark the test as completed successfully.")
  sonobuoy.passTest()

print("This is an example Starlark script")
foo()

print("You can set environment variables like SONOLARK_<NAME> and access them using env.<name>: " + env.foo)