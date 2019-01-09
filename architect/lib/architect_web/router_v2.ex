defmodule ArchitectWeb.V2Router do
  use ArchitectWeb, :router

  scope "/swagger" do
    forward("/", PhoenixSwagger.Plug.SwaggerUI,
      otp_app: :architect,
      swagger_file: "v2.swagger.json"
    )
  end

  def swagger_info do
    %{
      info: %{
        version: "2.0",
        title: "VelocityCI Architect"
      }
    }
  end

  pipeline :authenticated do
    plug(Architect.Pipelines.Guardian)
    plug(Guardian.Plug.EnsureAuthenticated)
  end

  pipeline :api do
    plug(:accepts, ["json"])

    plug(Plug.Parsers,
      parsers: [:urlencoded, :multipart, :json, Absinthe.Plug.Parser],
      pass: ["*/*"],
      json_decoder: Poison
    )
  end

  scope "/" do
    pipe_through(:api)

    forward("/graphiql", Absinthe.Plug.GraphiQL, schema: ArchitectWeb.Schema)
    forward("/", Absinthe.Plug, schema: ArchitectWeb.Schema)
  end
end
