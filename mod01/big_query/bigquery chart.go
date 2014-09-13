package big_query

// https://godoc.org/code.google.com/p/google-api-go-client/bigquery/v2
// https://developers.google.com/bigquery/bigquery-api-quickstart
import (

	"image"
	"image/color"
	"image/png"
	"os"
	"fmt"

	"net/http"
	"appengine"
	"github.com/pbberlin/tools/util_err"
	"github.com/pbberlin/tools/util"
	"github.com/pbberlin/tools/charting"
	"github.com/pbberlin/tools/adapter"


)




func showAsTable(w http.ResponseWriter, r *http.Request , m map[string]interface{} ){

	cd1 := GetChartDataFromDatastore(w,r,"chart_data_01")	
	cd := *cd1

	span := util.GetSpanner()
	// Header row
	fmt.Fprintf(w,span(" ",164)	)		
	for _,lg := range cd.VLangs {
		fmt.Fprintf(w,span(lg,88)	)		
	}
	fmt.Fprintf(w,"<br>")		
	
	for _,period := range cd.VPeriods {
		fmt.Fprintf(w,span(period,164)	)		
		for _,lg := range cd.VLangs {
			fmt.Fprintf(w,span( cd.M[period][lg]  ,88)	)		
		}
		fmt.Fprintf(w,"<br>")		
	}

	
}


func showAsChart(w http.ResponseWriter, r *http.Request, m map[string]interface{} ){

	cd1 := GetChartDataFromDatastore(w,r,"chart_data_01")	
	cd := *cd1


	c := appengine.NewContext(r)

	
	
	optScale,_,_:= charting.BestScale(cd.F_max, charting.Scale_y_vm)
	scale_max := 0.0
	for _,val := range optScale {
		//fmt.Fprintf(w,"%v - %v \n", tick, val)		
		fVal := util.Stof(val)
		if fVal > scale_max { scale_max = fVal }
	}


	
	p := r.FormValue("p")
	if p == "" { p = "static/chartbg_400x960__480x1040__12x10.png" }	

	f, err := os.Open(p)
	util_err.Err_http(w,r,err,false)
	defer f.Close()
	

	imgRaw, whichFormat, err := image.Decode(f)
	util_err.Err_http(w,r,err,false,"only jpeg and png are 'activated' ")
	c.Infof( "serving img format %v %T\n" , whichFormat , imgRaw)
	
	var img *image.RGBA
	img,ok := imgRaw.(*image.RGBA)
	util_err.Err_http(w,r,ok,false,"chart bg must have interal format RGBA")


	for langIndex,lang := range cd.VLangs {

		gci := langIndex % len(charting.GraphColors)  // graph color index
	
		lineCol := color.RGBA{charting.GraphColors[gci][0],
			charting.GraphColors[gci][1],
			charting.GraphColors[gci][2],
			0,
		}

		//fmt.Fprintf(w,"%v %v \n",gci,lineCol)

		
		drw := charting.FuncDrawLiner(lineCol,img)
		xb,yb := 40,440
		//P0 := image.Point{xb,yb}
		//drw( P0, lineCol,img ) 

		x,y := xb,yb

		maxPeriods := 0
		for _,period := range cd.VPeriods {

			tmp := cd.M[period][lang]/scale_max * 400
			y = yb - int( tmp )


			drw( image.Point{x,y}, lineCol,img) 			
			//fmt.Fprintf(w,"%v-%v: %v => %v => %v\n",period, lang,count,int(tmp),y)
			x += 40

			maxPeriods++
			if maxPeriods>24 {break}
		}
	}

	w.Header().Set("Content-Type", "image/png")
	png.Encode(w, img)


	charting.SaveImageToDatastore( w,r, img, "chart2" )	

}

func init() {

	http.HandleFunc("/big-query/show-chart", adapter.Adapter(showAsChart) )
	http.HandleFunc("/big-query/show-table", adapter.Adapter(showAsTable) )
}