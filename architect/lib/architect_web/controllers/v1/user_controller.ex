defmodule ArchitectWeb.V1.UserController do
  use ArchitectWeb, :controller
  use PhoenixSwagger

  alias Architect.Accounts
  alias Architect.Accounts.User
  # alias Architect.Accounts.Guardian

  action_fallback(ArchitectWeb.V1.FallbackController)

  def swagger_definitions do
    %{
      AuthRequest:
        swagger_schema do
          title("AuthRequest")
          description("POST body for authenticating a user")

          properties do
            username(:string, "Username", required: true)
            password(:string, "Password", required: true)
          end

          example(%{
            username: "bob",
            password: "bob's password"
          })
        end,
      AuthResponse:
        swagger_schema do
          title("AuthResponse")
          description("Response schema for authentication token")

          properties do
            token(:string, "Token")
            username(:string, "Username")
          end

          example(%{
            token: "",
            username: "bob"
          })
        end,
      User:
        swagger_schema do
          title("User")
          description("A user of Velocity CI")

          properties do
            id(:string, "UUID")
            username(:string, "Username", required: true)
          end

          example(%{
            id: "",
            username: "bob"
          })
        end,
      UserRequest:
        swagger_schema do
          title("UserRequest")
          description("POST body for creating a user")

          properties do
            username(:string, "Username", required: true)
            password(:string, "Password", required: true)
          end

          example(%{
            username: "bob",
            password: "bob's password"
          })
        end,
      UserResponse:
        swagger_schema do
          title("UserResponse")
          description("Response schema for single user")
          property(:data, Schema.ref(:User), "The user details")
        end,
      UsersResponse:
        swagger_schema do
          title("UsersReponse")
          description("Response schema for multiple users")
          property(:data, Schema.array(:User), "The users details")
        end
    }
  end

  swagger_path(:index) do
    get("/v1/users")
    summary("List users")
    description("List all users")
    produces("application/json")
    deprecated(false)

    response(200, "OK", Schema.ref(:UsersResponse))
  end

  def index(conn, _params) do
    users = Accounts.list_users()
    render(conn, "index.json", users: users)
  end

  swagger_path(:create) do
    post("/v1/users")
    summary("Create user")
    description("Creates a new User with the given credentials")
    produces("application/json")
    deprecated(false)
    ArchitectWeb.CommonParameters.authorization()

    parameter(:user, :body, Schema.ref(:UserRequest), "The user details",
      example: %{username: "bob", password: "bob's password"}
    )

    response(201, "Created", Schema.ref(:UserResponse))
    response(400, "Bad Request")
  end

  def create(conn, user_params) do
    with {:ok, %User{} = user} <- Accounts.create_user(user_params) do
      conn
      |> put_status(:created)
      |> put_resp_header("location", V1Routes.user_path(conn, :show, user))
      |> render("show.json", user: user)
    end
  end

  def delete(conn, %{"id" => id}) do
    user = Accounts.get_user!(id)

    with {:ok, %User{}} <- Accounts.delete_user(user) do
      send_resp(conn, :no_content, "")
    end
  end

  swagger_path(:auth_create) do
    post("/v1/auth")
    summary("Authenticate user")
    description("Creates a new authentication token for a user, given their correct credentials")
    produces("application/json")
    deprecated(false)

    parameter(:user, :body, Schema.ref(:UserRequest), "The user details",
      example: %{username: "bob", password: "bob's password"}
    )

    response(201, "Created", Schema.ref(:AuthResponse))
    response(401, "Unauthorized")
  end

  def auth_create(conn, user_params) do
    with %{params: %{"username" => username, "password" => password}} <-
           User.changeset(%User{}, user_params),
         {:ok, user} <- Accounts.authenticate(username, password) do
      {:ok, token, claims} = Accounts.encode_and_sign(user)

      conn
      |> render("auth.json", %{user: user, token: token, claims: claims})
    end
  end

  def auth_error(conn, {_type, _reason}, _opts) do
    conn
    |> put_view(ArchitectWeb.V1.UserView)
    |> put_status(:unauthorized)
    |> render("auth_error.json")
  end
end
