# GoGoGadgetGravity

GoGoGadgetGravity is a project created as a personal Go learning experience. It is meant to be fun, and nothing about it is perfect. There are minor issues with the correctness of the physics system, and insanely stupid scene drawing work around involving temporary image files, and undoubtedly plenty of non-idiomatic code (more than anything else, this is what I'd love feedback on / pull requests for, if anyone is so inclined).

GoGoGadgetGravity is a particle simulator, including physics engine and gui packages, which uses artificial physics:

Gravity is inversely proportional to distance^2.
- It is always positive and therefore attractive.
- Masses add. Radius is proxy.

Close Charge is inversely proportional to distance^3.
- It may be negative or positive and therefore repulsive or attractive
- Charges average. Red (negative) and green (positive) are proxy (zero is black), with charge min/max +/- 1.

Far Charge is *proportional* to distance.
- It is always positive and therefore attractive.
- Charges average. Alpha is proxy with charge range  0-1.


## Prerequisites

The Qt API by TheRecipe is required if using the guis\qt package. To install it:\
`set GO111MODULE=off`\
`go get -v github.com/therecipe/qt/cmd/... && for /f %v in ('go env GOPATH') do 
    %v\bin\qtsetup test && %v\bin\qtsetup -test=false`\
`set GO111MODULE=auto`\
The first time you build, a new folder "qtbox" will be created in the build directory, with the redistributable (platform dependent) Qt component.