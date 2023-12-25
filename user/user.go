// Verify and authenticates users with the database
// we will have different ty0pes of users
// admin users the users that can do anything and correct delete and update anything
// sellers users will be on charge of billing and receiving money
// delivery users will be the ones on sending products and transport goods
//
// each user will have different type of commands
// for example when a delivery user receive a shipping will be marked as on delivery
// when the goods are delivered we mark them as done
package user


// Base user that is the root os all the other user types
type user{
	
}
// delivery type user will be on charge of transport goods
type Delivery struct{

}

// se
type Seller struct{
	
}

type Admin struct {

}
