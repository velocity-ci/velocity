module Page.Errored exposing (view, pageLoadError, PageLoadError)

{-| The page that renders when there was an error trying to load another page,
for example a Page Not Found error.
-}

import Html exposing (..)
import Html.Attributes exposing (class, tabindex, id, alt)
import Data.Session as Session exposing (Session)
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


view : Session -> PageLoadError -> Html msg
view session (PageLoadError model) =
    div [ class "container mt-3" ]
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
