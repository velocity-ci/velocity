defmodule ArchitectWeb.Router do
  use ArchitectWeb, :router

  pipeline :api do
    plug(:accepts, ["json"])

    plug(Plug.Parsers,
      parsers: [Absinthe.Plug.Parser],
      pass: ["*/*"]
    )

    plug(ArchitectWeb.Context)
  end

  scope "/" do
    pipe_through(:api)

    if Mix.env() == :dev do
      forward("/graphiql", Absinthe.Plug.GraphiQL, schema: ArchitectWeb.Schema)
    end

    forward("/", Absinthe.Plug, schema: ArchitectWeb.Schema)
  end
end
