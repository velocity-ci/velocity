module Request.Project exposing (..)

import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.Project as Project exposing (Project)
import Json.Decode as Decode
import Request.Helpers exposing (apiUrl)
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Http


list : Maybe AuthToken -> Http.Request (List Project)
list maybeToken =
    let
        expect =
            Decode.list (Project.decoder)
                |> Http.expectJson
    in
        apiUrl "/projects"
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest
