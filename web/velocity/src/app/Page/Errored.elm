module Page.Errored exposing (PageLoadError, pageLoadError, view)

{-| The page that renders when there was an error trying to load another page,
for example a Page Not Found error.
-}

import Data.Session as Session exposing (Session)
import Html exposing (..)
import Html.Attributes exposing (alt, class, id, tabindex)
import Views.Page as Page exposing (ActivePage)


-- MODEL --


type PageLoadError
    = PageLoadError Model


type alias Model =
    { activePage : ActivePage
    , errorMessage : String
    }


pageLoadError : ActivePage -> String -> PageLoadError
pageLoadError activePage errorMessage =
    PageLoadError { activePage = activePage, errorMessage = errorMessage }



-- VIEW --


view : Session msg -> PageLoadError -> Html msg
view session (PageLoadError model) =
    div [ class "p-4 mt-3" ]
        [ div [ class "row" ]
            [ div [ class "col-12" ]
                [ h1 [ class "display-4 text-danger" ]
                    [ i [ class "fa fa-exclamation-triangle" ] []
                    , text " An error occurred."
                    ]
                ]
            ]
        , div [ class "row" ]
            [ div [ class "col-12" ]
                [ samp [] [ text model.errorMessage ] ]
            ]
        ]
