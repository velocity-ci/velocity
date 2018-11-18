module Project exposing (Project, decoder, list, name, repository, thumbnailSrc)

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint exposing (Endpoint)
import Http
import Iso8601
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import PaginatedList exposing (PaginatedList)
import Project.Id as Id exposing (Id)
import Project.Slug as Slug exposing (Slug)
import Time


type Project
    = Project Internals


type alias Internals =
    { id : Id
    , slug : Slug
    , name : String
    , repository : String
    , createdAt : Time.Posix
    , updatedAt : Time.Posix
    , synchronising : Bool
    , logo : Maybe String
    }



-- SERIALIZATION --


decoder : Decoder Project
decoder =
    Decode.succeed Project
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "id" Id.decoder
        |> required "slug" Slug.decoder
        |> required "name" Decode.string
        |> required "repository" Decode.string
        |> required "createdAt" Iso8601.decoder
        |> required "updatedAt" Iso8601.decoder
        |> required "synchronising" Decode.bool
        |> required "logo" (Decode.maybe Decode.string)



-- INFO --


name : Project -> String
name (Project project) =
    project.name


thumbnailSrc : Project -> Maybe String
thumbnailSrc (Project project) =
    project.logo


repository : Project -> String
repository (Project project) =
    project.repository



-- COLLECTION --


list : Maybe Cred -> BaseUrl -> Http.Request (List Project)
list maybeCred baseUrl =
    let
        endpoint =
            Endpoint.projects { amount = -1, page = 1 } (Api.toEndpoint baseUrl)
    in
    Decode.field "data" (Decode.list decoder)
        |> Api.get endpoint maybeCred
