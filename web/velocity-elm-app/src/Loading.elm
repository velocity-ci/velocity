module Loading exposing (error, icon, slowThreshold)

{-| A loading spinner icon.
-}

import Asset
import Element exposing (Element)
import Process
import Svg exposing (..)
import Svg.Attributes exposing (..)
import Task exposing (Task)


icon : { width : Float, height : Float } -> Element msg
icon conf =
    svg
        [ width (String.fromFloat conf.width)
        , height (String.fromFloat conf.height)
        , viewBox "0 0 58 58"
        ]
        [ loaderSvg ]
        |> Element.html


error : String -> Element msg
error str =
    Element.text ("Error loading " ++ str ++ ".")


slowThreshold : Task x ()
slowThreshold =
    Process.sleep 500



-- PRIVATE


loaderSvg : Svg msg
loaderSvg =
    let
        spinnerCircle conf =
            circle
                [ cx conf.cx
                , cy conf.cy
                , r "5"
                , fillOpacity conf.fillOpacity
                , fill "currentColor"
                ]
                [ animate
                    [ attributeName "fill-opacity"
                    , begin "0s"
                    , dur "1.3s"
                    , values conf.animateValues
                    , calcMode "linear"
                    , repeatCount "indefinite"
                    ]
                    []
                ]
    in
    g [ fill "none", fillRule "evenodd" ]
        [ g [ transform "translate(2 1)" ]
            [ spinnerCircle { cx = "42.601", cy = "11.462", fillOpacity = "1", animateValues = "1;0;0;0;0;0;0;0" }
            , spinnerCircle { cx = "49.063", cy = "27.063", fillOpacity = "0", animateValues = "0;1;0;0;0;0;0;0" }
            , spinnerCircle { cx = "42.601", cy = "42.663", fillOpacity = "0", animateValues = "0;0;1;0;0;0;0;0" }
            , spinnerCircle { cx = "27.000", cy = "49.125", fillOpacity = "0", animateValues = "0;0;0;1;0;0;0;0" }
            , spinnerCircle { cx = "11.399", cy = "42.663", fillOpacity = "0", animateValues = "0;0;0;0;1;0;0;0" }
            , spinnerCircle { cx = "4.9380", cy = "27.063", fillOpacity = "0", animateValues = "0;0;0;0;0;1;0;0" }
            , spinnerCircle { cx = "11.399", cy = "11.462", fillOpacity = "0", animateValues = "0;0;0;0;0;0;1;0" }
            , spinnerCircle { cx = "27.000", cy = "5.0000", fillOpacity = "0", animateValues = "0;0;0;0;0;0;0;1" }
            ]
        ]
