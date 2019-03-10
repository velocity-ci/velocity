module Form.Select exposing (select)

import Element exposing (Element)
import Element.Background as Background
import Element.Border as Border
import Element.Button as Button
import Element.Events exposing (onClick)
import Element.Font as Font
import Html exposing (..)


select : Element msg
select =
    Element.html <|
        Html.select [] []
