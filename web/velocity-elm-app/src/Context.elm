module Context exposing (Context, baseUrl, device, start, windowResize)

{-| The runtime context of the application.
-}

import Api exposing (BaseUrl, Cred)
import Element exposing (Device)


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


device : Context msg -> Device
device (Context _ val) =
    val



-- CHANGES


windowResize : { width : Int, height : Int } -> Context msg -> Context msg
windowResize dimensions (Context baseUrl_ _) =
    Context baseUrl_ (Element.classifyDevice dimensions)
