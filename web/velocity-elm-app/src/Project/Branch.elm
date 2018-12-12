module Project.Branch exposing (Branch, list, text)

import Api exposing (BaseUrl, Cred)
import Api.Endpoint as Endpoint exposing (Endpoint)
import Element exposing (..)
import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode
import Project.Branch.Name as Name exposing (Name)
import Project.Slug as ProjectSlug


type Branch
    = Branch Internals


type alias Internals =
    { name : Name
    , active : Bool
    }



-- INFO


text : Branch -> Element msg
text (Branch { name }) =
    Name.text name



-- SERIALIZATION --


decoder : Decoder Branch
decoder =
    Decode.succeed Branch
        |> custom internalsDecoder


internalsDecoder : Decoder Internals
internalsDecoder =
    Decode.succeed Internals
        |> required "name" Name.decoder
        |> required "active" Decode.bool



-- COLLECTION --


list : Cred -> BaseUrl -> ProjectSlug.Slug -> Http.Request (List Branch)
list cred baseUrl projectSlug =
    let
        endpoint =
            Endpoint.branches (Just { amount = -1, page = 1 }) (Api.toEndpoint baseUrl) projectSlug
    in
        Decode.field "data" (Decode.list decoder)
            |> Api.get endpoint (Just cred)
