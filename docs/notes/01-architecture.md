- [Hexagonal Architecture with Domain Driven Design (DDD)](#hexagonal-architecture-with-domain-driven-design-ddd)
  - [Hexagonal Architecture vs Onion Architecture](#hexagonal-architecture-vs-onion-architecture)
  - [`application/dto`](#applicationdto)
- [Resources](#resources)

# Hexagonal Architecture with Domain Driven Design (DDD)

**Hexagonal Architecture**, also known as **Ports and Adapters**, promotes decoupling by organizing an application around its domain, separating the core business logic from external concerns like UI, databases, and third-party systems. When combined with **Domain-Driven Design** (DDD), this architecture allows us to model complex business problems effectively and translate them into clean, maintainable code.

**DDD** helps us define the core domain using building blocks like:

- **Entities**: objects with identity
- **Value Objects**: immutable objects representing concepts
- **Aggregates**: clusters of Entities and Value Objects treated as a single unit
- **Repositories**: abstractions for data access
- **Domain Services**: operations that don't naturally belong to any Entity or Value Object.

This well-defined domain, residing at the **heart of the hexagon**, interacts with the outside world through ports (**interfaces**) and adapters (**implementations**), enabling flexibility, testability, and independent evolution of both the **domain and external dependencies**.

## Hexagonal Architecture vs Onion Architecture

The hexagons can be mapped almost one-to-one to the rings of the onion architecture:

- The “external agencies” arranged around the outer hexagon are represented in the onion architecture by the infrastructure components at the bottom right.
- The outer hexagon “adapters” corresponds to the ring containing “user interface,” “tests,” and “infrastructure.”
- The application hexagon corresponds to the application core in the onion architecture. This is further subdivided into “application services,” “domain services,” and “domain model,” whereby only the “domain model” is a fixed component of the onion architecture. The other rings of the application core are explicitly marked as optional. The “domain model” defines the “enterprise business rules” and thus corresponds to the “entities” ring – i.e., the innermost circle – of the clean architecture.

![hexagonal-architecture-vs-onion-architecture](/docs/images/hexagonal-architecture-vs-onion-architecture.png)

<center>Hexagonal architecture vs onion architecture</center>

## `application/dto`

**Purpose**: Its primary purpose is to **transfer data** between layers of your application or between your application and an external system (like a client via an API). It encapsulates a subset of the data from your BookDetail domain object.

**Reduced Data Exposure**: It deliberately exposes only the Title and AvailableCopies fields. This avoids exposing potentially sensitive or unnecessary data (like internal IDs, creation timestamps, or other fields that the client doesn't need to know about). This is a core principle of good API design and security.

**Decoupling**: It decouples your internal domain model (BookDetail) from the specific data contract of your API endpoint. You can change the structure of BookDetail internally without affecting your API clients, as long as the BookTitleAvailability struct and the transformation logic remain consistent. This reduces the risk of breaking changes.

**Serialization/Deserialization**: It's designed for easy serialization into JSON (or other formats) when sending data over the network and potentially for deserialization if you were receiving this structure from another service. The json:"..." tags control how the fields are mapped during (de)serialization.

**No Business Logic**: DTOs should not contain business logic. They are purely data containers. BookTitleAvailability adheres to this; it only holds data and has no methods that perform any logic.

# Resources

- [GeeksforGeeks - Hexagonal Architecture - System Design](https://www.geeksforgeeks.org/hexagonal-architecture-system-design/)
- [HappyCoders - Hexagonal Architecture - What is it? Why should you use it?](https://www.happycoders.eu/software-craftsmanship/hexagonal-architecture/)
- [Netflix Technology Blog - Ready for changes with Hexagonal Architecture](https://netflixtechblog.com/ready-for-changes-with-hexagonal-architecture-b315ec967749)
- [Redis - Domain Driven Design (DDD)](https://redis.io/glossary/domain-driven-design-ddd/)
