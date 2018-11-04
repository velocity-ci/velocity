module Context exposing (Context, baseUrl, fromBaseUrlAndDimensions, windowResize)

{-| The runtime context of the application.
-}

import Api exposing (BaseUrl)
import Element exposing (Device)
import Email exposing (Email)
import Json.Decode as Decode exposing (Decoder)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Encode as Encode exposing (Value)
import Username exposing (Username)



-- TYPES


type Context
    = Context BaseUrl Device



-- INFO


baseUrl : Context -> BaseUrl
baseUrl (Context val _) =
    val



-- CHANGES


windowResize : { width : Int, height : Int } -> Context -> Context
windowResize dimensions (Context baseUrl_ _) =
    Context baseUrl_ (Element.classifyDevice dimensions)


fromBaseUrlAndDimensions : BaseUrl -> { width : Int, height : Int } -> Context
fromBaseUrlAndDimensions baseUrl_ dimensions =
    Context baseUrl_ (Element.classifyDevice dimensions)
