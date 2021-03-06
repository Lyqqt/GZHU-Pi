const iconCamera = "https://shaw-1256261760.cos.ap-guangzhou.myqcloud.com/gzhu-pi/images/icon/xiangji.svg"
const avatar = "https://shaw-1256261760.cos.ap-guangzhou.myqcloud.com/gzhu-pi/images/icon/anonmous_avatar.png"
const tabs = [{
  name: "日常",
  placeholder: "标题 闲聊/求助/组队...",
  imageNum: 3,
  anonymous: false
}, {
  name: "情墙",
  placeholder: "表白对象",
  imageNum: 1,
  anonymous: false
}, {
  name: "悄悄话",
  placeholder: "标题 你的心声",
  imageNum: 3,
  anonymous: true
}, {
  name: "二手",
  placeholder: "标题 二手物品",
  imageNum: 3,
  anonymous: false
}]


Page({
  data: {
    mode: wx.$param["mode"],

    tabs: tabs,
    currentTab: 0,
    iconCamera: iconCamera,
    avatar: avatar,

    title: "",
    content: "",
    anonymous: false,
    anonymity: "",
    imgList: [],
    label: [],
    addi: {
      file_ids: [],
      contact: "",
      remark: ""
    },
  },

  fake() {
    let mode = wx.$param["mode"]
    this.setData({
      mode: mode
    })
    if (mode == "prod") {
      return false
    } else return true
  },

  onLoad: function (options) {

    if (this.fake()) return

    // 切换Tab
    let name = options.id
    let e = {
      detail: {
        index: 0
      }
    }
    for (let i in tabs) {
      if (tabs[i].name == name) {
        e.detail.index = i
      }
    }
    this.tabChange(e)

    // 当前用户
    wx.BaaS.auth.getCurrentUser().then(user => {
      console.log("user", user)
      // if (user.gender == 0) this.data.anonymity = "匿名童鞋"
      // if (user.gender == 1) this.data.anonymity = "匿名小哥哥"
      // if (user.gender == 2) this.data.anonymity = "匿名小姐姐"
      this.data.anonymity = user.nickname
      this.setData({
        anonymity: this.data.anonymity
      })
    }).catch(err => {
      this.setData({
        anonymity: "匿名童鞋"
      })
      if (err.code === 604) {
        console.log('用户未登录')
      }
    })
  },

  // true 说明在防抖期间，应该停止执行
  isDebounce(timeout = 2000) {
    let that = this
    if (this.data.debounce) {
      console.log("触发防抖")
      return true
    }
    this.data.debounce = true
    setTimeout(() => {
      that.data.debounce = false
    }, timeout)
    return false
  },

  post() {
    wx.$subscribe()
    if (this.isDebounce(5000)) return
    this.checkAndSave()
  },

  async checkAndSave() {

    let form = {
      type: this.data.tabs[this.data.currentTab].name,
      title: this.data.title,
      content: this.data.content,
      image: [],
      label: this.data.label,
      anonymous: this.data.anonymous,
      addi: this.data.addi,
    }

    if (form.anonymous) {
      form.anonymity = this.data.anonymity
      if (form.anonymity == "") {
        form.anonymity = "匿名童鞋"
      }
    }
    if (form.title == "" || form.content == "") {
      wx.showToast({
        title: '标题/内容不能为空',
        icon: "none",
      })
      return
    }

    // 违规检测
    let text = JSON.stringify(form)
    if (await wx.$checkText(text)) {
      return
    }

    this.setData({
      loading: true
    })

    if (this.data.imgList.length == 0) {
      this.saveRecord(form)
      return
    }

    // 批量上传图片
    let uploadTask = []
    for (let i = 0; i < this.data.imgList.length; i++) {
      uploadTask.push(this.uploadFile('', this.data.imgList[i]))
    }
    Promise.all(uploadTask).then((result) => {
      console.log("上传结果", result)
      for (let i in result) {
        form.image.push(result[i].path)
        form.addi.file_ids.push(result[i].file.id)
      }
      this.saveRecord(form)
    }).catch(err => {
      console.error(err)
    })
  },

  saveRecord(form = {}) {
    if (typeof form != "object") {
      console.error("error in form type")
      return
    }
    // prest 不能直接处理jsonb需要转字符串
    form.addi = JSON.stringify(form.addi)
    console.log("表单数据", form)
    // 保存数据
    wx.$ajax({
        url: wx.$param.server["prest"] + wx.$param.server["scheme"] + "/t_topic",
        // url: "http://localhost:9000/api/v1/postgres/public/t_topic",
        data: form,
        loading: true,
        header: {
          "content-type": "application/json"
        }
      })
      .then(res => {
        this.setData({
          loading: false
        })
        console.log("发布结果：", res.data)
        if (res.statusCode >= 400) {
          wx.showModal({
            title: '发布失败',
            content: res.errMsg + res.data ? res.data.msg : "",
          })
          return
        }
        if (typeof res.error != "undefined") {
          wx.showModal({
            title: '发布失败',
            content: res.error,
          })
          return
        }
        // TODO 错误判断
        wx.showToast({
          title: '发布成功',
        })
        setTimeout(function () {
          wx.redirectTo({
            url: '/pages/Life/wall/detail?id=' + res.data.id,
          })
        }, 1000)
      }).catch(err => {
        console.log(err)
        this.setData({
          loading: false
        })
      })
  },


  // 添加标签
  labelAdd: function () {
    if (this.data.labelInput == "" || this.data.labelInput == undefined) {
      wx.showToast({
        title: '请先输入标签内容',
        icon: "none"
      })
      return
    }
    if (this.data.label.length == 3) {
      wx.showToast({
        title: '最多加3个标签',
        icon: "none"
      })
      return
    }
    this.setData({
      label: this.data.label.concat(this.data.labelInput),
      labelInput: ""
    })
  },

  // 删除标签
  labelDel(e) {
    let id = Number(e.target.id)
    this.data.label.splice(id, 1)
    this.setData({
      label: this.data.label
    })
  },

  // 读取标签内容
  labelInput: function (e) {
    this.data.labelInput = e.detail.value
  },

  /*
   * 伪双向绑定
   * wxml input 定义属性：data-field="field1.field2" value="{{field1.field2}}"
   * 输入内容将绑定到：this.data.field1.field2 = e.detail.value
   */
  inputBind(e) {
    if (typeof e.currentTarget.dataset.field != "string") return
    let field = e.currentTarget.dataset.field
    // console.log("数据绑定：key：", field, " value:", e.detail.value)

    let data = {}
    data[field] = e.detail.value
    this.setData(data)
  },

  tabChange(e) {
    if (this.data.tabs[e.detail.index].name == "二手") {
      wx.$navTo("/pages/Life/oldthings/post")
      return
    }

    // 切换tab，删除多余的图片
    if (this.data.imgList.length >= this.data.tabs[e.detail.index].imageNum) {
      this.data.imgList = this.data.imgList.slice(0, this.data.tabs[e.detail.index].imageNum)
    }

    this.setData({
      currentTab: e.detail.index,
      anonymous: this.data.tabs[e.detail.index].anonymous,
      imgList: this.data.imgList,
      hideAddBtn: this.data.imgList.length >= this.data.tabs[e.detail.index].imageNum ? true : false,
      curImgIndex: 0,
    })
  },

  anonymousSwitch(e) {
    this.setData({
      anonymous: e.detail.value
    })
    if (e.detail.value) {
      let profile_pic = wx.getStorageSync('gzhupi_user').profile_pic
      if (typeof (profile_pic) == "string" && profile_pic != "") {
        this.setData({
          avatar: profile_pic
        })
      }
    }
  },
  tapMore() {
    this.setData({
      openMore: this.data.openMore ? false : true
    })
  },


  // 异步上传单个文件
  uploadFile: function (categoryName, filePath) {
    let MyFile = new wx.BaaS.File()
    let metaData = {
      categoryName: categoryName
    }
    //返回上传文件后的信息
    return new Promise(function (callback) {
      let fileParams = {
        filePath: filePath
      }
      MyFile.upload(fileParams, metaData).then(res => {
        callback(res.data)
      }, err => {
        wx.showToast({
          title: '上传失败',
          icon: "none"
        })
      })
    })
  },

  chooseImage() {
    var maxImageNum = this.data.tabs[this.data.currentTab].imageNum
    wx.chooseImage({
      count: maxImageNum - this.data.imgList.length, //默认9
      sizeType: ['compressed'],
      sourceType: ['album', "camera"],
      success: (res) => {
        this.checkImage(res.tempFilePaths)
        var imgList = this.data.imgList.concat(res.tempFilePaths)
        this.setData({
          imgList: imgList,
          curImgIndex: imgList.length - 1,
          hideAddBtn: imgList.length >= this.data.tabs[this.data.currentTab].imageNum ? true : false
        })
      }
    });
  },

  checkImage(tempFilePaths = []) {
    for (let i = 0; i < tempFilePaths.length; i++) {
      wx.BaaS.wxCensorImage(tempFilePaths[i]).then(res => {
        console.log("图片检测", res.risky)
        if (res.risky) {
          wx.showModal({
            title: '警告',
            content: '您的发布图片包含违规内容',
          })
          this.setData({
            imgList: []
          })
        }
      }, err => {
        // HError 对象
      })
    }

  },

  viewImage(e) {
    wx.previewImage({
      urls: this.data.imgList,
      current: e.currentTarget.dataset.url
    });
  },

  deleteImage(e) {
    console.log(e)
    wx.showModal({
      title: '提示',
      content: '确定要删除这张照片吗？',
      cancelText: '留着',
      confirmText: '再见',
      success: res => {
        if (res.confirm) {
          this.data.imgList.splice(e.currentTarget.dataset.index, 1);
          this.setData({
            imgList: this.data.imgList,
            curImgIndex: (e.currentTarget.dataset.index - 1) < 0 ? 0 : (e.currentTarget.dataset.index - 1),
            hideAddBtn: this.data.imgList.length >= this.data.tabs[this.data.currentTab].imageNum ? true : false
          })
        }
      }
    })
  },

})