module Route exposing (Route(..), fromUrl, link, replaceUrl)

import Browser.Navigation as Nav
import Element exposing (Attribute, Element)
import Url exposing (Url)
import Url.Parser as Parser exposing ((</>), Parser, oneOf, s, string)
import Username exposing (Username)



-- ROUTING


type Route
    = Home
    | Root
    | Login
    | Logout


parser : Parser (Route -> a) a
parser =
    oneOf
        [ Parser.map Home Parser.top
        , Parser.map Login (s "login")
        , Parser.map Logout (s "logout")
        ]



-- PUBLIC HELPERS


link : List (Attribute msg) -> Element msg -> Route -> Element msg
link attrs label targetRoute =
    Element.link attrs
        { url = routeToString targetRoute
        , label = label
        }


replaceUrl : Nav.Key -> Route -> Cmd msg
replaceUrl key route =
    Nav.replaceUrl key (routeToString route)


fromUrl : Url -> Maybe Route
fromUrl url =
    Parser.parse parser url



-- INTERNAL


routeToString : Route -> String
routeToString page =
    let
        pieces =
            case page of
                Home ->
                    []

                Root ->
                    []

                Login ->
                    [ "login" ]

                Logout ->
                    [ "logout" ]
    in
    "/" ++ String.join "/" pieces
