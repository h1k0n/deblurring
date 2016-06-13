package main

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/jpeg"
	"log"
	"net/http"
	"os"
	"text/template"
	"fmt"
	"github.com/runningwild/go-fftw/fftw"
	_ "golang.org/x/image/bmp"
	"image/color"
	"image/png"
    _"image/gif"
	"math"
	"encoding/json"
    "strconv"
)
//Result holds the JSON RETURN
type Result struct{
    Dataurl string
}
//Pargs is form post arguments
type Pargs struct {
	mode string
	method string
	radius int
	sigma float64
	direction float64
}
var templates = template.Must(template.ParseFiles("templates/index.html", "templates/show.html"))
//IndexHandler handle '/' deprecated
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]interface{}{"Title": "index"}
	renderTemplate(w, "index", data)
}
//UploadHandler handle '/upload'
func UploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Allowed POST method only", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(32 << 20) // maxMemory
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	file, _, err := r.FormFile("upload")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()
    img, _, err := image.Decode(file)
    if err!=nil{
        http.Error(w, err.Error(), http.StatusInternalServerError)
		return
    }
	f, err := os.Create("/tmp/test.png")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer f.Close()
    err=png.Encode(f,img)
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//io.Copy(f, file)
	//http.Redirect(w, r, "/show", http.StatusFound)
}
//ShowHandler handle '/show'
func ShowHandler(w http.ResponseWriter, r *http.Request) {
    err:=r.ParseForm()
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	mode:=r.PostFormValue("mode")
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
    method:=r.PostFormValue("method")
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	radius,err:= strconv.Atoi(r.PostFormValue("radius"))
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	direction,err:= strconv.ParseFloat(r.PostFormValue("direction"),64)
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	sigma,err:= strconv.ParseFloat( r.PostFormValue("sigma"),64)
    if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	pargs:=Pargs{mode,method,radius,sigma,direction}
    
    fmt.Println(mode,method,radius,pargs)
    
	file, err := os.Open("/tmp/test.png")
	w.Header().Set("content-type", "application/json")
	defer file.Close()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	img, _, err := image.Decode(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	str:=writeImageWithTemplate(w, "show", &img,&pargs)
	out := &Result{str}
    	b, err := json.Marshal(out)
    	if err != nil {
        	return
    	}
    	w.Write(b)
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	if err := templates.ExecuteTemplate(w, tmpl+".html", data); err != nil {
		log.Fatalln("Unable to execute template.")
	}
}

// Array2RGB 图像三原色的复数数组
type Array2RGB [3]*fftw.Array2

//NewArray2RGB 产生Array2RGB
func NewArray2RGB(height, width int) Array2RGB {
	var rgb Array2RGB
	for i := 0; i < 3; i++ {
		rgb[i] = fftw.NewArray2(height, width)
	}
	return rgb
}

func initRGB(r, g, b complex128, i, j int, rgb Array2RGB) {
	rgb[0].Set(i, j, r)
	rgb[1].Set(i, j, g)
	rgb[2].Set(i, j, b)
}

func rgbFFT(rgb Array2RGB) Array2RGB {
	var rgbfft Array2RGB
	rgbfft[0] = fftw.FFT2(rgb[0])
	rgbfft[1] = fftw.FFT2(rgb[1])
	rgbfft[2] = fftw.FFT2(rgb[2])
	return rgbfft

}

func rgbIFFT(rgb Array2RGB) Array2RGB {
	var rgbifft Array2RGB
	rgbifft[0] = fftw.IFFT2(rgb[0])
	rgbifft[1] = fftw.IFFT2(rgb[1])
	rgbifft[2] = fftw.IFFT2(rgb[2])
	return rgbifft

}

func multiArray(c complex128, slice []complex128) []complex128 {
	length := len(slice)
	for i := 0; i < length; i++ {
		slice[i] = slice[i] * c
	}
	return slice
}

func multiArrayArr(a, b []complex128) []complex128 {
	L := len(a)
	if len(a) < len(b) {

		L = len(b)
	}
	c := make([]complex128, L)
	for i := 0; i < L; i++ {
		c[i] = a[i] * b[i]
	}
	return c
}

func multiRGB(c []complex128, rgb Array2RGB) Array2RGB {
	newRGB := NewArray2RGB(rgb[0].Dims())
	for i := 0; i < 3; i++ {
		newRGB[i].Elems = multiArrayArr(c, rgb[i].Elems)
	}
	return newRGB
}


func writeImageWithTemplate(w http.ResponseWriter, tmpl string, img *image.Image,pargs *Pargs) string{
	buffer := new(bytes.Buffer)
	var blur string
	if pargs.mode=="gaussian" {
		blur="GaussianBlur"
	}else {
		blur="MotionBlur"
	}
    var method string
    if pargs.method=="wiener" {
        method="Wiener"
    }else {
        method="LeastSquare"
    }
	
	radius:=pargs.radius
	sigma:=pargs.sigma
	dire:=pargs.direction
	

	//////////////////////////////////////////////////////////////////////////
	width, height := (*img).Bounds().Dx(), (*img).Bounds().Dy()
	length := width * height //照片的总像素数

	//为了和SmartDeblur一致，将(width,height)调整为(height,width)
	rgb := NewArray2RGB(height, width) //Array2RGB的大小为width,height
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			q, w, e, _ := (*img).At(x, y).RGBA() //q,w,e表示rgb分量的颜色值
			initRGB(complex(float64(q>>8), 0), complex(float64(w>>8), 0), complex(float64(e>>8), 0), y, x, rgb)
		}
	}

	
	var kernel []float64
	if blur == "GaussianBlur" {
		kernel = getGaussianKernel(radius, sigma)
	} else {
		kernel = getMotionKernel(radius, dire)
	}
	///////////////////////////////////////////
	//file3,err:=os.Create("kernelMatrix.txt")
	kernelMatrix := fftw.NewArray2(height, width) //模糊核
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {

			if math.Abs(float64(x-width/2)) <= float64(radius) && math.Abs(float64(y-height/2)) <= float64(radius) {
				xLocal := x - (width/2 - radius)
				yLocal := y - (height/2 - radius)
				kernelMatrix.Set(y, x, complex(kernel[yLocal*(2*radius+1)+xLocal], 0))
			}
		}
	}

	KernelTempMatrix := fftw.NewArray2(height, width)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			xTranslated := (x + width/2) % width
			yTranslated := (y + height/2) % height
			//kernelMatrix.Elems[y*width + x] = kernelMatrix.Elems[yTranslated*width + xTranslated]
			//这样KernelMatirx会被破坏
			// if x == 0 && y == 0 {
			// 	print(xTranslated, yTranslated, kernelMatrix.At(yTranslated, xTranslated))
			// 	print(xTranslated, yTranslated, kernelMatrix.Elems[yTranslated*width+xTranslated])
			// }
			KernelTempMatrix.Elems[y*width+x] = kernelMatrix.Elems[yTranslated*width+xTranslated]
			//fmt.Fprintf(file3,"[%d][%d]:%v ",x,y,kernelTempMatrix.At(x,y))
		}
		//fmt.Fprintln(file3)//0
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			kernelMatrix.Elems[y*width+x] = KernelTempMatrix.Elems[y*width+x]
		}
	}


	kernelFFT := fftw.FFT2(kernelMatrix)
	outTemp := fftw.NewArray2(height, width)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			outTemp.Elems[y*width+x] = complex(real(kernelFFT.Elems[y*width+x]), real(kernelFFT.Elems[y*width+x]))
		}
	}
	fftw.IFFT2To(outTemp, outTemp)
	///
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			index := y*width + x
			if x < 11 || y < 11 || x > width-11 || y > height-11 {
				realPart := real(outTemp.Elems[y*width+x]) / float64(width*height)
				imagPart := imag(outTemp.Elems[y*width+x]) / float64(width*height)
				complexTemp := complex(realPart, imagPart)
				rgb[0].Elems[index] = complexTemp
				rgb[1].Elems[index] = complexTemp
				rgb[2].Elems[index] = complexTemp
			}

		}
	}
	rgbfft := rgbFFT(rgb)

	if method == "Wiener" {
		// ///////////////////////////////////////////////////////////////////
		fmt.Println("维纳滤波开始")
		K := 0.0007
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				energyValue := math.Pow(real(kernelFFT.At(y, x)), 2) + math.Pow(imag(kernelFFT.At(y, x)), 2)
				wienerValue := complex(real(kernelFFT.At(y, x))/(energyValue+K), 0)
				rgbfft[0].Elems[y*width+x] = rgbfft[0].Elems[y*width+x] * wienerValue
				rgbfft[1].Elems[y*width+x] = rgbfft[1].Elems[y*width+x] * wienerValue
				rgbfft[2].Elems[y*width+x] = rgbfft[2].Elems[y*width+x] * wienerValue
			}
		}
		fmt.Println("维纳滤波完成")
		//////////////////////////////////////////////////////
	} else {
		/////////////////////////////////////////////////////////////////////
		fmt.Println("最小二乘滤波开始")
		laplacianMatrix := fftw.NewArray2(height, width)
		laplacianMatrix.Set(0, 0, complex(4, 0))
		laplacianMatrix.Set(0, 1, complex(-1, 0))
		laplacianMatrix.Set(0, width-1, complex(-1, 0))
		laplacianMatrix.Set(1, 0, complex(-1, 0))
		laplacianMatrix.Set(height-1, 0, complex(-1, 0))

		laplacianMatrixFFT := fftw.FFT2(laplacianMatrix)
		K := 0.007
		//////////////////////////////////////////////
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				energyValue := math.Pow(real(kernelFFT.At(y, x)), 2) + math.Pow(imag(kernelFFT.At(y, x)), 2)
				energyLaplacianValue := math.Pow(real(laplacianMatrixFFT.At(y, x)), 2) + math.Pow(imag(laplacianMatrixFFT.At(y, x)), 2)
				tikhonovValue := complex(real(kernelFFT.At(y, x))/(energyValue+K*energyLaplacianValue), 0)
				rgbfft[0].Elems[y*width+x] = rgbfft[0].Elems[y*width+x] * tikhonovValue
				rgbfft[1].Elems[y*width+x] = rgbfft[1].Elems[y*width+x] * tikhonovValue
				rgbfft[2].Elems[y*width+x] = rgbfft[2].Elems[y*width+x] * tikhonovValue
			}
		}
		fmt.Println("最小二乘滤波完成")
		//////////////////////////////////////////////////////
	}

	rgbifft := rgbIFFT(rgbfft)

	file2, err := os.Create("out.png")
	defer file2.Close()
	if err != nil {
		log.Fatal(err)
	}

	rgba := image.NewRGBA((*img).Bounds())
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			a := real(rgbifft[0].At(y, x)) / float64(length)
			b := real(rgbifft[1].At(y, x)) / float64(length)
			c := real(rgbifft[2].At(y, x)) / float64(length)
			if a > 255 {
				//print(a)
				a = 255

			}
			if b > 255 {
				//print(b)
				b = 255
			}
			if c > 255 {
				//print(c)
				c = 255
			}
			if a < 0 {
				a = 0
			}
			if b < 0 {
				b = 0
			}
			if c < 0 {
				c = 0
			}
			rgba.Set(x, y, color.RGBA{
				uint8(a), //因为除以了255,导致图片不清晰，导致折腾太久，搞得我以为要四舍五入
				uint8(b),
				uint8(c),
				255,
			})

		}
	}
	//////////////////////////////////////////////////////////////////////////



	if err := jpeg.Encode(buffer, rgba, nil); err != nil {
		log.Fatalln("Unable to encode image.")
	}

	str := base64.StdEncoding.EncodeToString(buffer.Bytes())
	return str
}

func main() {
	http.Handle("/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/upload", UploadHandler)
	http.HandleFunc("/show", ShowHandler)
	http.ListenAndServe(":8888", nil)
}


func getGaussianKernel(radius int, sigma float64) []float64 {
	width := 2*radius + 1
	kernel := make([]float64, width*width)
	xishu := 1.0 / (2.0 * math.Pi * sigma * sigma)
	xishu2 := -1.0 / (2.0 * sigma * sigma)
	for i := 0; i < width; i++ {
		for j := 0; j < width; j++ {
			zhishu := float64(((i-radius)*(i-radius) + (j-radius)*(j-radius))) * xishu2
			kernel[i*width+j] = xishu * math.Exp(zhishu)
		}
	}
	var sum float64
	for i := 0; i < width*width; i++ {
		sum += kernel[i]
	}
	for i := 0; i < width*width; i++ {
		kernel[i] /= sum
	}
	return kernel
}

func getMotionKernel(radius int, jiaodu float64) []float64 {
	width := 2*radius + 1
	kernel := make([]float64, width*width)
	//计算横纵坐标，往横纵坐标塞值
	//先用特殊情况对待45度，0度
	kernel[radius*width+radius] = 1
	//计算一四象限
	for x := 1; x <= radius; x++ {
		//计算纵坐标
		y := int(float64(x) * jiaodu)
		index := (x + radius) + width*(radius-y)
		if index > width*width || index < 0 {
			kernel[width*(radius-x)+radius] = 1
		} else {
			kernel[index] = 1
		}

	}
	//计算二三象限
	for x := -1; x >= -radius; x-- {
		//计算纵坐标
		y := int(float64(x) * jiaodu)
		index := (x + radius) + width*(radius-y)
		if index > width*width || index < 0 {
			kernel[width*(radius-x)+radius] = 1
		} else {
			kernel[index] = 1
		}
	}
	var sum float64
	for i := 0; i < width*width; i++ {
		sum += kernel[i]
	}
	for i := 0; i < width*width; i++ {
		kernel[i] /= sum
	}
	return kernel
}