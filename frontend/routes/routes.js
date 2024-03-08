const express = require('express');
const router = express.Router();
const axios = require('axios');
const multer = require('multer');
const { LocalStorage } = require('node-localstorage');
const localStorage = new LocalStorage('./scratch');

const storage = multer.diskStorage({
    destination: function (req, file, cb) {
        cb(null, 'uploads/') 
    },
    filename: function (req, file, cb) {
        cb(null, file.originalname)
    }
});
const upload = multer({ 
    storage: storage,
    limits: { fileSize: 5 * 1024 * 1024 }, 
}).single('image');

router.get('/', (req, res) => {
    console.log("start / ")
    res.status(200).render('register');
});

router.post('/register', async (req, res)=>{
    console.log("start register")
    const {email, username}= req.body
    axios.post('http://localhost:8080/getvcode', { email })
    .then(response => {
        if (response.status == 200) {
            res.redirect(`/code?email=${email}&username=${username}`)
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
})
router.get('/code', async(req, res)=>{
    console.log("start code get ")
    const username = req.query.username
    const email = req.query.email
    console.log(username, email)
    res.render('code', {username, email})
})
router.post('/code', async(req, res)=>{
    console.log("start code post")

    const username = req.query.username
    const email = req.query.email

    console.log(username, email)
    const code = req.body.code
    axios.post('http://localhost:8080/checkvcode', {
        
            email: email,
            code: code
        
    })
    .then(response => {
        if (response.status == 200) {
            res.redirect(`/getpass?email=${email}&username=${username}`)
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
})
router.get('/getpass', (req, res)=>{
    console.log("start getpass get")

    const username = req.query.username
    const email = req.query.email
    res.render('getpass', {username, email})
})
router.post('/getpass', (req, res)=>{
    console.log("start getpass post")

    const username = req.query.username
    const email = req.query.email
    const password = req.body.password
    axios.post('http://localhost:8080/register', {
            email: email,
            username: username, 
            password: password
        
    })
    .then(response => {
        if (response.status == 200) {
            const token = response.data.token;
            localStorage.setItem('token', token);
            res.redirect(`/login`)
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
})


router.get('/login', (req, res) => {
    console.log("start login get")

    res.status(200).render('login');
});

router.post('/login', async (req, res)=>{
    console.log("start login post")

    const {email, username, password}= req.body
    axios.post('http://localhost:8080/login', { email, username, password })
    .then(response => {
        const token = response.data.token;
        localStorage.setItem('token', token);

        if (response.status == 200) {
            res.redirect(`index`)
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
})

router.get('/index', (req, res) => {
    console.log("start index")

    const token = localStorage.getItem('token');

    var { order, sort_by, page } = req.query; 

    if (order==undefined){
        order ="asc"
    }
    if (sort_by == undefined){
        sort_by = "name"
    }
    if (page == undefined){
        page = 1
    }

    axios.get(`http://localhost:8080/courses?order=${order}&sort_by=${sort_by}&page=${page}`, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
        .then(response => {
            if (response.status == 200) {
                const sortBy = req.query.sortBy || 'defaultSortBy'; // Default value if sortBy is not provided
                res.render(`index`, {
                    courses: response.data,
                    currentPage: 1, // Добавьте текущую страницу в зависимости от запроса
                    totalPages: 5, // Здесь должна быть логика для определения общего количества страниц
                    sortBy: sortBy
                });
            }
        })
        .catch(error => {
            console.error('Error:', error);
        });
});

router.get("/courses/:courseId", (req, res)=>{
    console.log("start /:courseId")

    const token = localStorage.getItem('token');
    const courseId = req.params.courseId
    axios.get(`http://localhost:8080/courses/${courseId}`, {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (response.status == 200) {
            res.render(`course`, { course: response.data.course });
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
})
router.get("/enroll/:courseId", (req, res) => {
    console.log("start enroll")

    const token = localStorage.getItem('token');
    const courseId = req.params.courseId;
    
    axios.post(`http://localhost:8080/courses/${courseId}`, {},{
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        console.log(response.data)
        if (response.status === 200) {
            res.redirect(`/mycourses`);
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
});

router.get('/profile', (req, res)=>{
    console.log("start profile")

    const token = localStorage.getItem('token');
    axios.get('http://localhost:8080/profile', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (response.status == 200) {
            console.log(response.data.user)
            res.render(`profile`, {user: response.data.user})
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
})

router.get('/logout', (req, res)=>{
    res.redirect("/")
})

router.post('/update', (req, res) => {
    try {
        console.log("start update user get")
        const token = localStorage.getItem('token');
        const { username, email } = req.body;
        axios.post('http://localhost:8080/update',{username, email}, {
            headers: {
                'Authorization': `Bearer ${token}`
            }
        })
        .then(response => {
            if (response.status == 200) {
                res.redirect(`/profile`)
            }
        })
        .catch(error => {
            console.error('Error:', error);
            res.status(500).send('Internal Server Error');
        });
    } catch (error) {
        console.error('Error updating user information:', error);
        res.status(500).send('Internal Server Error');
    }
});

router.get('/subscribe', (req, res)=>{
    console.log("start subscribe get")
    const token = localStorage.getItem('token');
    axios.get('http://localhost:8080/subscribe', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (response.status == 200 || response.status == 400) {
            console.log(response.data)
            res.redirect('/index')
        }
      
    })
    .catch(error => {
        console.error('Error:', error);
    });
})

router.post('/admin/spam', (req, res)=>{
    console.log("start subscribe get")
    const text = req.body.text
    console.log(text)
    const token = localStorage.getItem('token');
    axios.post('http://localhost:8080/admin/send',{text},{
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (response.status == 200 || response.status == 400) {
            console.log(response.data)
            res.redirect('/admin')
        }
      
    })
    .catch(error => {
        console.error('Error:', error);
    });
})
router.get('/mycourses', (req,res)=>{
    console.log("start mycourses get")
    const token = localStorage.getItem('token');
    axios.get('http://localhost:8080/courses/my', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (response.status == 200) {
            console.log(response.data)
            res.render(`mycourses`, {courses: response.data.courses})
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
})
router.get('/admin', (req, res)=>{
    console.log("start admin get")
    const token = localStorage.getItem('token');
    axios.get('http://localhost:8080/all', {
        headers: {
            'Authorization': `Bearer ${token}`
        }
    })
    .then(response => {
        if (response.status == 200) {
            res.render(`admin`, {courses: response.data})
        }
    })
    .catch(error => {
        console.error('Error:', error);
    });
})


router.post("/admin",upload, (req, res) => {
    console.log("start admin create")

    const token = localStorage.getItem('token');
    axios.post('http://localhost:8080/courses/create',{
        name: req.body.name,
        description: req.body.description,
        image: req.file.path,
        duration: req.body.duration,
        price: req.body.price,
        places: req.body.places,
        category: req.body.category,
    }, {
        headers: {
            'Authorization': `Bearer ${token}`, 
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (response.status == 200) {
            res.redirect(`admin`)
        }
    })
    .catch(error => {
        console.error('Error:', error);
        res.status(500).json({ error: 'Ошибка сервера' });
    });
});

router.post('/admin/delete/:courseId', (req, res)=>{
    console.log("start admin dleete post")

    const token = localStorage.getItem('token');
    const courseId = req.params.courseId
    axios.delete(`http://localhost:8080/courses/${courseId}`, {
        headers: {
            'Authorization': `Bearer ${token}`, 
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (response.status == 200) {
            res.redirect(`/admin`)
        }
    })
    .catch(error => {
        console.error('Error:', error);
        res.status(500).json({ error: 'Ошибка сервера' });
    });
})
router.post("/admin/update/:courseId", (req, res) => {
    console.log("start admin update")

    const token = localStorage.getItem('token');
    const courseId = req.params.courseId
    axios.put(`http://localhost:8080/courses/${courseId}`,{
        name: req.body.name,
        description: req.body.description,
        duration: req.body.duration,
        price: req.body.price,
        places: req.body.places,
        category: req.body.category,
    }, {
        headers: {
            'Authorization': `Bearer ${token}`, 
            'Content-Type': 'application/json'
        }
    })
    .then(response => {
        if (response.status == 200) {
            res.redirect(`/admin`)
        }
    })
    .catch(error => {
        console.error('Error:', error);
        res.status(500).json({ error: 'Ошибка сервера' });
    });
});


module.exports = router;

