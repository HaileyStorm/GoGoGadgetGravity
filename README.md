# GoGoGadgetGravity

GoGoGadgetGravity is a project created as a personal Go learning experience. It is meant to be fun, and nothing about it is perfect. There are minor issues with the correctness of the physics system, and insanely stupid scene drawing work around involving temporary image files, and undoubtedly plenty of non-idiomatic code (more than anything else, this is what I'd love feedback on / pull requests for, if anyone is so inclined).

GoGoGadgetGravity is a particle simulator, including physics engine and gui packages, which uses artificial physics:
Gravity is inversely proportional to distance^2.
  It is always positive and therefore attractive.
  Masses add. Radius is proxy.
Close Charge is inversely proportional to distance^3.
  It may be negative or positive and therefore repulsive or attractive
  Charges average. Red (negative) and green (positive) are proxy (zero is black), with charge min/max +/- 1.
Far Charge is *proportional* to distance.
  It is always positive and therefore attractive.
  Charges average. Alpha is proxy with charge range  0-1.
