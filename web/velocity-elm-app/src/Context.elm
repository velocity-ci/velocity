module Context exposing (Context, baseUrl, device, start, windowResize, wsUrl)

{-| The runtime context of the application.
-}

import Api exposing (BaseUrl, Cred)
import Dict exposing (Dict)
import Element exposing (Device)
import Email exposing (Email)
import Http
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode exposing (Value)
import Project exposing (Project)
import Task exposing (Task)
import Username exposing (Username)


-- TYPES


type Context msg
    = Context BaseUrl Device


start : BaseUrl -> { width : Int, height : Int } -> Context msg
start baseUrl_ dimensions =
    let
        device_ =
            Element.classifyDevice dimensions
    in
        Context baseUrl_ device_



-- INFO


baseUrl : Context msg -> BaseUrl
baseUrl (Context val _) =
    val


wsUrl : Context msg -> String
wsUrl (Context val _) =
    Api.toWsEndpoint val ++ "/v1/ws"


device : Context msg -> Device
device (Context _ val) =
    val



-- CHANGES


windowResize : { width : Int, height : Int } -> Context msg -> Context msg
windowResize dimensions (Context baseUrl_ _) =
    Context baseUrl_ (Element.classifyDevice dimensions)
