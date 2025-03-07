# Why does Go favors explicit naming and composition over method overloading?

## Clarity and Readability

**Explicit is better than implicit**: This is a core tenet of Go. Method overloading relies on the compiler to determine which method is called based on the types and number of arguments. This implicit behavior can make code harder to understand at a glance, particularly for complex function signatures or when dealing with type conversions and implicit interfaces. With explicit naming and composition, the intent is always clear because the caller explicitly chooses the specific method to invoke.

Reduced mental burden: When reading code with overloaded methods, you need to consider all possible types and combinations of arguments to be sure which method will be executed. This adds to the mental burden of understanding the code. Explicit names make the flow of execution more obvious.

## Simplicity and Compile Time Resolution

**Simpler compiler**: Implementing method overloading adds complexity to the compiler, as it has to perform type checking and resolution to determine the correct method to call. Go's designers aimed for a smaller, faster compiler. By avoiding overloading, the compiler can resolve method calls more quickly and efficiently.

Faster compilation: Method overloading needs more runtime checking and decision-making. Explicit naming makes the decision-making process faster and prevents any ambiguity.

## Maintainability and Refactoring

**Easier to change code**: When refactoring code with overloaded methods, changes in argument types can have cascading effects, potentially breaking other parts of the codebase by unexpectedly resolving to a different overloaded method. With explicit naming, such changes are more localized and easier track.

Reduces the risk of shadowing: Overloading can lead to situations where a method in a derived class with a similar name accidentally shadows a method in a base class, leading to unexpected behavior. Explicit naming helps prevent these kinds of accidental conflicts.

## Composition Over Inheritance

**Go favors composition**: Go leans heavily towards composition as a mechanism for code reuse rather than inheritance. Composition allows you to embed types within other types, providing access to their methods. In this context, method overloading is less necessary as you are composing your types with their methods.

Interfaces as a mechanism for polymorphism: Go uses interfaces to achieve polymorphism. By explicitly defining the methods that an interface requires, the code becomes more reliable and explicit.

## Error Prevention

**Reduced ambiguity**: Explicit naming reduces the risk of accidentally calling the wrong method due to subtle type conversions or ambiguity in argument matching. The caller is directly responsible for choosing the right method to invoke.

## Example

```go
type BookService interface {
	GetBookByTitleHandler(c *fiber.Ctx) error
	GetBookByTitle(title string) (*entity.BookDetail, *apperrors.RestErr)
}
```

Explicit Naming (`GetBookByTitleHandler` vs. `GetBookByTitle`): Instead of relying on overloading (which is absent!), the code defines two distinct methods with very specific names: `GetBookByTitleHandler` and `GetBookByTitle`. This signals clear separation of concerns.

- `GetBookByTitleHandler`: This likely handles the web request processing directly (taking a fiber.Ctx context), extracting the title from the request, and then passing it on. The "Handler" suffix is a common naming convention that explicitly says "this function handles a specific request."
- `GetBookByTitle`: This likely contains the core business logic for retrieving a book by title from a data source (e.g., a database). It receives the title string and returns the data itself (`*entity.BookDetail`). This explicit separation avoids a single, overloaded method trying to do both request handling and database retrieval, making the code far more modular and easier to understand.

Clear Error Handling (Returns `*apperrors.RestErr`): The `GetBookByTitle` method returns a specific error type, `*apperrors.RestErr`. The `GetBookByTitleHandler` returns a general error type. It clearly states the potential types of errors that can occur, promoting robust error handling. Go avoids exceptions commonly used in other languages and utilizes standard error handling, increasing application predictability.
