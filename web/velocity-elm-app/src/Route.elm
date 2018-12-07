module Route exposing (Route(..), fromUrl, link, replaceUrl)

import Browser.Navigation as Nav
import Element exposing (..)
import Element.Font as Font
import Page.Home.ActivePanel as ActivePanel exposing (ActivePanel)
import Palette
import Project.Build.Id as BuildId
import Project.Id as ProjectId
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
    | Project ProjectId.Id



--    | Build BuildId.Id


parser : Parser (Route -> a) a
parser =
    oneOf
        [ Parser.map Home (Parser.top <?> ActivePanel.queryParser)
        , Parser.map Login (s "login")
        , Parser.map Logout (s "logout")
        , Parser.map Project (s "project" </> ProjectId.urlParser)
        ]



-- PUBLIC HELPERS


link : List (Attribute msg) -> Element msg -> Route -> Element msg
link attrs label targetRoute =
    Element.link
        (List.concat
            [ linkAttrs
            , attrs
            ]
        )
        { url = routeToString targetRoute
        , label = label
        }


linkAttrs : List (Attribute msg)
linkAttrs =
    [ Font.color Palette.primary3
    , mouseOver
        [ Font.color Palette.primary5
        ]
    ]


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

        Project id ->
            ( [ "project", ProjectId.toString id ], [] )


routeToString : Route -> String
routeToString page =
    let
        ( urlPieces, queryPieces ) =
            routePieces page
    in
    "/" ++ String.join "/" urlPieces ++ Url.Builder.toQuery queryPieces
