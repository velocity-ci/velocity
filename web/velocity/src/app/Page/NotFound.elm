module Page.NotFound exposing (view)

import Data.Session as Session exposing (Session)
import Html exposing (Html, div, h1, img, main_, text)
import Html.Attributes exposing (alt, class, id, src, tabindex)


-- VIEW --


view : Session msg -> Html msg
view session =
    main_ [ id "content", class "p-2", tabindex -1 ]
        [ h1 [] [ text "Not Found" ]
        , div [ class "row" ]
            [ text "Not found" ]
        ]
