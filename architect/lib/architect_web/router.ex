defmodule ArchitectWeb.Router do
  use ArchitectWeb, :router

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
