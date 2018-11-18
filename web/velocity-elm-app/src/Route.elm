module Route exposing (Route(..), fromUrl, link, replaceUrl)

import Browser.Navigation as Nav
import Element exposing (Attribute, Element)
import Page.Home.ActivePanel as ActivePanel exposing (ActivePanel)
import Url exposing (Url)
import Url.Builder exposing (QueryParameter)
import Url.Parser as Parser exposing ((</>), (<?>), Parser, oneOf, s, string)
import Url.Parser.Query as Query
import Username exposing (Username)



-- ROUTING


type Route
    = Home ActivePanel
    | Root
    | Login
    | Logout


parser : Parser (Route -> a) a
parser =
    oneOf
        [ Parser.map Home (Parser.top <?> ActivePanel.queryParser)
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


routePieces : Route -> ( List String, List QueryParameter )
routePieces page =
    case page of
        Home activePanel ->
            ( []
            , ActivePanel.toQueryParams activePanel
            )

        Root ->
            ( [], [] )

        Login ->
            ( [ "login" ], [] )

        Logout ->
            ( [ "logout" ], [] )


routeToString : Route -> String
routeToString page =
    let
        ( urlPieces, queryPieces ) =
            routePieces page
    in
    "/" ++ String.join "/" urlPieces ++ Url.Builder.toQuery queryPieces
