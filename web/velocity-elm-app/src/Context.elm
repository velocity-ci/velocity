module Context exposing (Context, baseUrl, fromBaseUrl)

{-| The runtime context of the application.
-}

import Api exposing (BaseUrl)
import Email exposing (Email)
import Json.Decode.Pipeline exposing (custom, required)
import Json.Decode as Decode exposing (Decoder)
import Json.Encode as Encode exposing (Value)
import Username exposing (Username)


-- TYPES


type Context
    = Context BaseUrl



-- INFO


baseUrl : Context -> BaseUrl
baseUrl (Context val) =
    val



-- CHANGES


fromBaseUrl : BaseUrl -> Context
fromBaseUrl =
    Context
