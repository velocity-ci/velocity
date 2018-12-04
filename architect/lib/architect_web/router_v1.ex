defmodule ArchitectWeb.V1Router do
  use ArchitectWeb, :router

  scope "/swagger" do
    forward "/", PhoenixSwagger.Plug.SwaggerUI,
      otp_app: :architect,
      swagger_file: "v1.swagger.json"
  end

  def swagger_info do
    %{
      info: %{
        version: "1.0",
        title: "VelocityCI Architect"
      }
    }
  end

  pipeline :authenticated do
    plug Architect.Pipelines.Guardian
    plug Guardian.Plug.EnsureAuthenticated
  end

  pipeline :api do
    plug(:accepts, ["json"])
  end

  scope "/", ArchitectWeb.V1 do
    pipe_through(:api)

    get "/health", HealthController, :index

    post "/auth", UserController, :auth_create
  end

  scope "/", ArchitectWeb.V1 do
    pipe_through([:authenticated, :api])
    resources("/users", UserController)
    resources("/ssh/known-hosts", KnownHostController)
  end
end
