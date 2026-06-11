// Package omit provides a type which can be used to represent a value which may or may not be set.
// This is useful for omitting the value in JSON if it can be null or omitted since pointers can't represent this.
//
// The zero value of Omit is not set.
//
// Example:
//
// 	type user struct {
// 		ID   Omit[int]     `json:"id,omitzero"`
// 		Name Omit[*string] `json:"name,omitzero"`
// 	}
//
// 	json.Marshal(user{
// 		ID:   New(1),
// 		Name: NewPtr("john"),
// 	})
//
// 	// Output: {"id":1,"name":"john"}
//
// 	json.Marshal(user{
// 		ID:   NewZero[int](),
// 		Name: NewNilPtr[string](),
// 	})
//
// 	// Output: {"name":null}
//
// 	var u user
// 	json.Unmarshal([]byte(`{"id":1,"name":"john"}`), &u)
// 	fmt.Println(u.ID.Value, u.Name.Value)
//
// 	// Output: 1 john
//
// 	json.Unmarshal([]byte(`{"name":null}`), &u)
// 	fmt.Println(u.ID, u.Name)
//
// 	// Output: <omitted> <nil>
//

package omit
