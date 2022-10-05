// Hi there! This is a comment.

Console println:"Hello"+"World".

sayHi <- [self|
	self println:"Hi!"]
	extends Console.

Console sayHi:.

myObject <- copy Console.
foo <- 42 extends myObject.

myObject println:(myObject foo:).