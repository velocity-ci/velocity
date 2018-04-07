module Request.KnownHost exposing (list, create)

import Context exposing (Context)
import Data.AuthToken as AuthToken exposing (AuthToken, withAuthorization)
import Data.KnownHost as KnownHost exposing (KnownHost)
import Json.Encode as Encode
import Request.Helpers exposing (apiUrl)
import Request.Errors
import HttpBuilder exposing (RequestBuilder, withBody, withExpect, withQueryParams)
import Util exposing ((=>))
import Http
import Data.PaginatedList as PaginatedList exposing (PaginatedList)
import Task exposing (Task)


baseUrl : String
baseUrl =
    "/ssh/known-hosts"



-- LIST --


list : Context -> Maybe AuthToken -> Task Request.Errors.HttpError (PaginatedList KnownHost)
list context maybeToken =
    let
        expect =
            KnownHost.decoder
                |> PaginatedList.decoder
                |> Http.expectJson
    in
        apiUrl context (baseUrl ++ "?amount=-1")
            |> HttpBuilder.get
            |> HttpBuilder.withExpect expect
            |> withAuthorization maybeToken
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleHttpError



-- CREATE --


type alias CreateConfig record =
    { record
        | scannedKey : String
    }


create : Context -> CreateConfig record -> AuthToken -> Task Request.Errors.HttpError KnownHost
create context config token =
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
        apiUrl context baseUrl
            |> HttpBuilder.post
            |> withAuthorization (Just token)
            |> withBody body
            |> withExpect expect
            |> HttpBuilder.toTask
            |> Task.mapError Request.Errors.handleHttpError
