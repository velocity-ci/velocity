module Request.KnownHost exposing (list, create)

import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.KnownHost as KnownHost exposing (KnownHost)
import Json.Encode as Encode
import Request.Helpers exposing (apiUrl)
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Util exposing ((=>))
import Http
import Data.PaginatedList as PaginatedList exposing (PaginatedList)


baseUrl : String
baseUrl =
    "/ssh/known-hosts"



-- LIST --


list : Maybe AuthToken -> Http.Request (PaginatedList KnownHost)
list maybeToken =
    let
        expect =
            KnownHost.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl baseUrl
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toRequest



-- CREATE --


type alias CreateConfig record =
    { record
        | scannedKey : String
    }


create : CreateConfig record -> AuthToken -> Http.Request KnownHost
create config token =
    let
        expect =
            KnownHost.decoder
                |> Http.expectJson

        project =
            Encode.object
                [ "entry" => Encode.string config.scannedKey ]

        body =
            project
                |> Http.jsonBody
    in
        apiUrl baseUrl
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withBody body
            |> withExpect expect
            |> HttpBuilder.toRequest
